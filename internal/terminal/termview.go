package terminal

import (
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

var (
	scrollTrackColor = lipgloss.Color("#374151")
	scrollThumbColor = lipgloss.Color("#7C3AED")
)

type OutputMsg struct {
	ID   string
	Data []byte
}

type TermErrorMsg struct {
	ID  string
	Err error
}

type cursorState struct {
	visible bool
}

type TermViewModel struct {
	id                string
	pty               *os.File
	emu               *vt.Emulator
	width             int
	height            int
	done              bool
	scrollOffset      int
	prevScrollbackLen int
	passthrough       bool
	cursorState       *cursorState
}

func NewTermViewModel(id string, ptyFile *os.File) TermViewModel {
	emu := vt.NewEmulator(80, 24)
	cs := &cursorState{}
	emu.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			cs.visible = visible
		},
	})
	// The emulator writes encoded key presses and replies to terminal
	// queries (cursor position, device attributes, ...) to its input
	// buffer; pump them into the pty so the child program receives them.
	go io.Copy(ptyFile, emu) //nolint:errcheck
	return TermViewModel{id: id, pty: ptyFile, emu: emu, cursorState: cs}
}

func (m TermViewModel) ID() string {
	return m.id
}

func (m TermViewModel) Done() bool {
	return m.done
}

func (m TermViewModel) Close() {
	m.emu.Close()
}

func (m TermViewModel) SetPassthrough(b bool) TermViewModel {
	m.passthrough = b
	return m
}

func (m TermViewModel) readCmd() tea.Cmd {
	ptyFile, id := m.pty, m.id
	return func() tea.Msg {
		buf := make([]byte, 32*1024)
		n, err := ptyFile.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			return OutputMsg{ID: id, Data: data}
		}
		if err != nil {
			return TermErrorMsg{ID: id, Err: err}
		}
		return OutputMsg{ID: id}
	}
}

func (m TermViewModel) Init() tea.Cmd {
	return m.readCmd()
}

func (m TermViewModel) Update(msg tea.Msg) (TermViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case OutputMsg:
		if msg.ID != m.id || m.done {
			return m, nil
		}
		if len(msg.Data) > 0 {
			oldSBLen := m.emu.ScrollbackLen()
			m.emu.Write(msg.Data) //nolint:errcheck
			newSBLen := m.emu.ScrollbackLen()
			delta := newSBLen - oldSBLen
			if m.scrollOffset > 0 && delta > 0 {
				m.scrollOffset += delta
			}
			m.prevScrollbackLen = newSBLen
			if m.scrollOffset > newSBLen {
				m.scrollOffset = newSBLen
			}
		}
		return m, m.readCmd()

	case TermErrorMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.done = true
		return m, nil

	case tea.KeyMsg:
		if !m.emu.IsAltScreen() {
			switch msg.Type {
			case tea.KeyCtrlU:
				return m.ScrollUp(m.height / 2), nil
			case tea.KeyCtrlD:
				return m.ScrollDown(m.height / 2), nil
			}
			if m.scrollOffset > 0 {
				m.scrollOffset = 0
			}
		}
		if m.passthrough {
			m.sendKey(msg)
		}
		return m, nil

	case tea.MouseMsg:
		if m.passthrough {
			m.sendMouse(msg)
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			return m.ScrollUp(3), nil
		case tea.MouseButtonWheelDown:
			return m.ScrollDown(3), nil
		}
		return m, nil
	}

	return m, nil
}

func (m TermViewModel) sendKey(msg tea.KeyMsg) {
	if m.done {
		return
	}
	if msg.Type == tea.KeyRunes && !msg.Alt {
		if msg.Paste {
			m.emu.Paste(string(msg.Runes))
		} else {
			m.emu.SendText(string(msg.Runes))
		}
		return
	}
	if key, ok := keyMsgToUV(msg); ok {
		m.emu.SendKey(key)
	}
}

func (m TermViewModel) sendMouse(msg tea.MouseMsg) {
	if m.done {
		return
	}
	m.emu.SendMouse(mouseMsgToVT(msg))
}

func mouseMsgToVT(msg tea.MouseMsg) vt.Mouse {
	button := uv.MouseButton(msg.Button)
	var mod uv.KeyMod
	if msg.Shift {
		mod |= uv.ModShift
	}
	if msg.Alt {
		mod |= uv.ModAlt
	}
	if msg.Ctrl {
		mod |= uv.ModCtrl
	}

	m := uv.Mouse{X: msg.X, Y: msg.Y, Button: button, Mod: mod}

	switch msg.Action {
	case tea.MouseActionPress:
		if button == vt.MouseWheelUp || button == vt.MouseWheelDown ||
			button == vt.MouseWheelLeft || button == vt.MouseWheelRight {
			return vt.MouseWheel(m)
		}
		return vt.MouseClick(m)
	case tea.MouseActionRelease:
		return vt.MouseRelease(m)
	case tea.MouseActionMotion:
		return vt.MouseMotion(m)
	}
	return vt.MouseClick(m)
}

func (m TermViewModel) SetSize(w, h int) TermViewModel {
	if w <= 0 || h <= 0 || (w == m.width && h == m.height) {
		return m
	}
	m.width = w
	m.height = h
	m.emu.Resize(w, h)
	ResizePTY(m.pty, uint16(h), uint16(w))
	if maxOff := m.emu.ScrollbackLen(); m.scrollOffset > maxOff {
		m.scrollOffset = maxOff
	}
	return m
}

func (m TermViewModel) ScrollUp(n int) TermViewModel {
	m.scrollOffset += n
	if maxOff := m.emu.ScrollbackLen(); m.scrollOffset > maxOff {
		m.scrollOffset = maxOff
	}
	return m
}

func (m TermViewModel) ScrollDown(n int) TermViewModel {
	m.scrollOffset -= n
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	return m
}

func (m TermViewModel) ScrollToBottom() TermViewModel {
	m.scrollOffset = 0
	return m
}

func (m TermViewModel) IsScrolledUp() bool {
	return m.scrollOffset > 0
}

func (m TermViewModel) View() string {
	var saved *uv.Cell
	var cx, cy int
	if m.cursorState.visible && m.scrollOffset == 0 {
		pos := m.emu.CursorPosition()
		cx, cy = pos.X, pos.Y
		if cx >= 0 && cy >= 0 && cx < m.width && cy < m.height {
			if orig := m.emu.CellAt(cx, cy); orig != nil {
				saved = orig.Clone()
				cell := orig.Clone()
				cell.Style.Fg, cell.Style.Bg = orig.Style.Bg, orig.Style.Fg
				m.emu.SetCell(cx, cy, cell)
			}
		}
	}

	var result string
	if m.emu.IsAltScreen() || (m.scrollOffset == 0 && m.emu.ScrollbackLen() == 0) {
		result = strings.TrimRight(m.emu.Render(), "\n")
	} else if m.scrollOffset == 0 {
		result = m.overlayScrollbar(strings.TrimRight(m.emu.Render(), "\n"))
	} else {
		result = m.renderScrolled()
	}

	if saved != nil {
		m.emu.SetCell(cx, cy, saved)
	}

	return result
}

func (m TermViewModel) renderScrolled() string {
	sb := m.emu.Scrollback()
	sbLen := sb.Len()
	screenLines := strings.Split(m.emu.Render(), "\n")

	totalLines := sbLen + len(screenLines)
	viewStart := totalLines - m.height - m.scrollOffset
	if viewStart < 0 {
		viewStart = 0
	}

	visible := make([]string, m.height)
	for i := range m.height {
		lineIdx := viewStart + i
		if lineIdx < sbLen {
			rendered := sb.Line(lineIdx).Render()
			visible[i] = rendered
		} else {
			screenIdx := lineIdx - sbLen
			if screenIdx >= 0 && screenIdx < len(screenLines) {
				visible[i] = screenLines[screenIdx]
			}
		}
	}

	return m.overlayScrollbar(strings.Join(visible, "\n"))
}

func (m TermViewModel) overlayScrollbar(content string) string {
	sb := m.emu.Scrollback()
	if sb == nil || sb.Len() == 0 || m.width <= 1 || m.height <= 0 {
		return content
	}

	totalLines := sb.Len() + m.height
	viewStart := totalLines - m.height - m.scrollOffset
	if viewStart < 0 {
		viewStart = 0
	}

	thumbSize := max(1, m.height*m.height/totalLines)
	scrollRange := max(1, totalLines-m.height)
	thumbPos := viewStart * (m.height - thumbSize) / scrollRange

	trackStyle := lipgloss.NewStyle().Foreground(scrollTrackColor)
	thumbStyle := lipgloss.NewStyle().Foreground(scrollThumbColor)

	lines := strings.Split(content, "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	contentWidth := m.width - 1
	for i := range lines {
		if i >= m.height {
			break
		}
		var glyph string
		if i >= thumbPos && i < thumbPos+thumbSize {
			glyph = thumbStyle.Render("┃")
		} else {
			glyph = trackStyle.Render("│")
		}
		lines[i] = ansi.Truncate(lines[i], contentWidth, "") + glyph
	}

	return strings.Join(lines[:min(len(lines), m.height)], "\n")
}

func ResizePTY(f *os.File, rows, cols uint16) error {
	return pty.Setsize(f, &pty.Winsize{Rows: rows, Cols: cols})
}
