package terminal

import (
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

type OutputMsg struct {
	ID   string
	Data []byte
}

type TermErrorMsg struct {
	ID  string
	Err error
}

type TermViewModel struct {
	id     string
	pty    *os.File
	emu    *vt.Emulator
	width  int
	height int
	done   bool
}

func NewTermViewModel(id string, ptyFile *os.File) TermViewModel {
	emu := vt.NewEmulator(80, 24)
	// The emulator writes encoded key presses and replies to terminal
	// queries (cursor position, device attributes, ...) to its input
	// buffer; pump them into the pty so the child program receives them.
	go io.Copy(ptyFile, emu) //nolint:errcheck
	return TermViewModel{id: id, pty: ptyFile, emu: emu}
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
			m.emu.Write(msg.Data) //nolint:errcheck
		}
		return m, m.readCmd()

	case TermErrorMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.done = true
		return m, nil

	case tea.KeyMsg:
		m.sendKey(msg)
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

func (m TermViewModel) SetSize(w, h int) TermViewModel {
	if w <= 0 || h <= 0 || (w == m.width && h == m.height) {
		return m
	}
	m.width = w
	m.height = h
	m.emu.Resize(w, h)
	ResizePTY(m.pty, uint16(h), uint16(w))
	return m
}

func (m TermViewModel) View() string {
	return strings.TrimRight(m.emu.Render(), "\n")
}

func ResizePTY(f *os.File, rows, cols uint16) error {
	return pty.Setsize(f, &pty.Winsize{Rows: rows, Cols: cols})
}
