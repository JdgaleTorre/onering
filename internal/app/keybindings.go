package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/JdgaleTorre/onering/internal/ui"
)

type KeyMap struct {
	Quit              key.Binding
	Help              key.Binding
	FocusLeft         key.Binding
	FocusRight        key.Binding
	Up                key.Binding
	Down              key.Binding
	Insert            key.Binding
	Escape            key.Binding
	Enter             key.Binding
	NewSession        key.Binding
	Delete            key.Binding
	Editor            key.Binding
	Git               key.Binding
	Docker            key.Binding
	Tab               key.Binding
	PageUp            key.Binding
	PassthroughEscape key.Binding
	Section1          key.Binding
	Section2          key.Binding
	Section3          key.Binding
	Info              key.Binding
	Refresh           key.Binding
	RecursiveRefresh  key.Binding
	PTYRun            key.Binding
	Favorite          key.Binding
	ToggleSidebar     key.Binding
	ExpandRight       key.Binding
	ExpandLeft        key.Binding
	ExpandRightLg     key.Binding
	ExpandLeftLg      key.Binding
	ToggleLayout      key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit:              key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Help:              key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	FocusLeft:         key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/left", "focus left")),
	FocusRight:        key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/right", "focus right")),
	Up:                key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "up")),
	Down:              key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "down")),
	Insert:            key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "insert/passthrough")),
	Escape:            key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "navigation mode")),
	Enter:             key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select/activate")),
	NewSession:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new session")),
	Delete:            key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete/kill")),
	Editor:            key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "open editor")),
	Git:               key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "open lazygit")),
	Docker:            key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "open lazydocker")),
	Tab:               key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle focus")),
	PageUp:            key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "page up")),
	PassthroughEscape: key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "escape passthrough")),
	Section1:          key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "jump to sessions")),
	Section2:          key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "jump to apps")),
	Section3:          key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "jump to tasks")),
	Info:              key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "project info")),
	Refresh:           key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh tasks")),
	RecursiveRefresh:  key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "recursive scan")),
	PTYRun:            key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "run task in PTY")),
	Favorite:          key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "toggle preferred")),
	ToggleSidebar:     key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "toggle sidebar")),
	ExpandRight:       key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "expand right")),
	ExpandLeft:        key.NewBinding(key.WithKeys("["), key.WithHelp("[", "expand left")),
	ExpandRightLg:     key.NewBinding(key.WithKeys("}"), key.WithHelp("}", "expand right large")),
	ExpandLeftLg:      key.NewBinding(key.WithKeys("{"), key.WithHelp("{", "expand left large")),
	ToggleLayout:      key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "toggle layout")),
}

func (k KeyMap) ImportantBindingGroups() []ui.BindingGroup {
	return []ui.BindingGroup{
		{
			Name: "navigation",
			Bindings: []key.Binding{
				k.FocusLeft, k.FocusRight,
				k.Up, k.Down,
			},
		},
		{
			Name: "sessions",
			Bindings: []key.Binding{
				k.NewSession, k.Delete, k.Insert, k.Enter,
			},
		},
		{
			Name: "apps",
			Bindings: []key.Binding{
				k.Editor, k.Git, k.Docker,
			},
		},
		{
			Name: "tasks",
			Bindings: []key.Binding{
				k.Section3, k.Refresh, k.RecursiveRefresh, k.PTYRun, k.Favorite,
			},
		},
		{
			Name: "layout",
			Bindings: []key.Binding{
				k.ToggleSidebar, k.ToggleLayout,
				k.ExpandLeft, k.ExpandRight,
				k.ExpandLeftLg, k.ExpandRightLg,
			},
		},
		{
			Name: "general",
			Bindings: []key.Binding{
				k.Help, k.Quit,
			},
		},
	}
}

func (k KeyMap) NavigationBindings() []key.Binding {
	return []key.Binding{
		k.Quit, k.Help, k.FocusLeft, k.FocusRight,
		k.Up, k.Down, k.Insert, k.Enter, k.NewSession, k.Delete,
		k.Editor, k.Git, k.Docker,
		k.Tab, k.PageUp,
		k.Section1, k.Section2, k.Section3, k.Info,
		k.Refresh, k.RecursiveRefresh, k.PTYRun, k.Favorite,
		k.ToggleSidebar, k.ToggleLayout,
		k.ExpandRight, k.ExpandLeft,
		k.ExpandRightLg, k.ExpandLeftLg,
	}
}

func (k KeyMap) InsertBindings() []key.Binding {
	return []key.Binding{
		k.Escape, k.Enter,
	}
}

func (k KeyMap) PassthroughBindings() []key.Binding {
	return []key.Binding{
		k.PassthroughEscape,
	}
}
