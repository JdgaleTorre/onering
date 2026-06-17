package terminal

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

var (
	colorSchemeCh   chan struct{}
	colorSchemeOnce sync.Once
	noopCh          = make(chan struct{})
)

func initColorSchemeWatcher() {
	colorSchemeCh = make(chan struct{}, 1)
	go watchXDGColorScheme()
}

func watchXDGColorScheme() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		colorSchemeCh = nil
		return
	}

	rule := "type='signal'," +
		"sender='org.freedesktop.portal.Desktop'," +
		"interface='org.freedesktop.portal.Settings'," +
		"member='SettingChanged'," +
		"path='/org/freedesktop/portal/desktop'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)

	signals := make(chan *dbus.Signal, 10)
	conn.Signal(signals)

	for sig := range signals {
		if len(sig.Body) < 2 {
			continue
		}
		ns, _ := sig.Body[0].(string)
		key, _ := sig.Body[1].(string)
		if ns == "org.freedesktop.appearance" && key == "color-scheme" {
			select {
			case colorSchemeCh <- struct{}{}:
			default:
			}
		}
	}
}

// ListenColorSchemeChange returns a tea.Cmd that blocks until the system
// color scheme changes (via XDG Desktop Portal). On systems without D-Bus
// it blocks forever (harmless no-op goroutine).
func ListenColorSchemeChange() tea.Cmd {
	colorSchemeOnce.Do(initColorSchemeWatcher)
	return func() tea.Msg {
		ch := colorSchemeCh
		if ch == nil {
			ch = noopCh
		}
		<-ch
		RedetectHostColors()
		return ColorSchemeChangedMsg{}
	}
}
