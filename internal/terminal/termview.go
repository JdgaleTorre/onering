package terminal

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

const (
	defaultTermWidth  = 80
	defaultTermHeight = 24
	ptyReadBufSize    = 32 * 1024
)

var (
	scrollTrackColor lipgloss.TerminalColor = lipgloss.ANSIColor(8) // Bright Black
	scrollThumbColor lipgloss.TerminalColor = lipgloss.ANSIColor(5) // Magenta
)

type OutputMsg struct {
	ID            string
	Data          []byte
	ScrollbackLen int
}

type ClearToastMsg struct {
	ID string
}

type TermErrorMsg struct {
	ID  string
	Err error
}

type cursorState struct {
	visible bool
	style   vt.CursorStyle
	blink   bool
}

type textSelection struct {
	active bool
	startX int
	startY int
	endX   int
	endY   int
}

func (s textSelection) normalized() (startX, startY, endX, endY int) {
	if s.startY < s.endY || (s.startY == s.endY && s.startX <= s.endX) {
		return s.startX, s.startY, s.endX, s.endY
	}
	return s.endX, s.endY, s.startX, s.startY
}

func (s textSelection) empty() bool {
	return s.startX == s.endX && s.startY == s.endY
}

type TermViewModel struct {
	id                string
	pty               *os.File
	emu               *vt.SafeEmulator
	width             int
	height            int
	done              bool
	scrollOffset      int
	prevScrollbackLen int
	passthrough       bool
	cursorState       *cursorState
	selection         textSelection
	childMouseMode    *atomic.Int32
	copyToast         bool
	mouseForwardMode  bool
}

func NewTermViewModel(id string, ptyFile *os.File) TermViewModel {
	emu := vt.NewSafeEmulator(defaultTermWidth, defaultTermHeight)
	applyHostColors(emu, getHostColors())
	cs := &cursorState{visible: true}
	mouseMode := &atomic.Int32{}
	emu.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			cs.visible = visible
		},
		CursorStyle: func(style vt.CursorStyle, blink bool) {
			cs.style = style
			cs.blink = blink
		},
		EnableMode: func(mode ansi.Mode) {
			if isMouseTrackingMode(mode) {
				mouseMode.Add(1)
			}
		},
		DisableMode: func(mode ansi.Mode) {
			if isMouseTrackingMode(mode) {
				if mouseMode.Load() > 0 {
					mouseMode.Add(-1)
				}
			}
		},
	})
	// Pump the emulator's reply stream (query responses, key echoes) back to
	// the child. Read directly from the underlying Emulator to avoid a
	// deadlock: SafeEmulator.Write holds a write lock while the OSC handler
	// blocks on the internal io.Pipe; routing io.Copy through
	// SafeEmulator.Read would need a read lock. Emulator.Read only touches the
	// pipe, which is already goroutine-safe.
	go io.Copy(ptyFile, emu.Emulator) //nolint:errcheck
	return TermViewModel{id: id, pty: ptyFile, emu: emu, cursorState: cs, childMouseMode: mouseMode}
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
	if b {
		m.selection = textSelection{}
	}
	return m
}

func (m TermViewModel) SetMouseForwardMode(b bool) TermViewModel {
	m.mouseForwardMode = b
	return m
}

// readCmd reads from the PTY and feeds data directly into the emulator so that
// terminal query responses (OSC 10/11, DA, etc.) are written back to the child
// process immediately, without waiting for a bubbletea Update round-trip.
func (m TermViewModel) readCmd() tea.Cmd {
	ptyFile, id, emu := m.pty, m.id, m.emu
	return func() tea.Msg {
		buf := make([]byte, ptyReadBufSize)
		n, err := ptyFile.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			// Answer OSC 4 palette queries ourselves: the VT emulator has no
			// OSC 4 handler, so without this the child's palette query goes
			// unanswered and it falls back to its default theme.
			respondPaletteQueries(ptyFile, data)
			emu.Write(data) //nolint:errcheck
			return OutputMsg{ID: id, Data: data, ScrollbackLen: emu.ScrollbackLen()}
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
			newSBLen := msg.ScrollbackLen
			delta := newSBLen - m.prevScrollbackLen
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

	case ClearToastMsg:
		if msg.ID == m.id {
			m.copyToast = false
			m.selection = textSelection{}
		}
		return m, nil

	case tea.KeyMsg:
		if !m.emu.IsAltScreen() {
			switch msg.Type {
			case tea.KeyCtrlK:
				return m.ScrollUp(m.height / 2), nil
			case tea.KeyCtrlJ:
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
		if m.passthrough && m.mouseForwardMode {
			m.sendMouse(msg)
			return m, nil
		}
		if msg.Button == tea.MouseButtonLeft {
			switch msg.Action {
			case tea.MouseActionPress:
				m.copyToast = false
				m.selection = textSelection{
					active: true,
					startX: msg.X, startY: msg.Y,
					endX: msg.X, endY: msg.Y,
				}
				return m, nil
			case tea.MouseActionMotion:
				if m.selection.active {
					m.selection.endX = msg.X
					m.selection.endY = msg.Y
					return m, nil
				}
			case tea.MouseActionRelease:
				if m.selection.active {
					m.selection.active = false
					m.selection.endX = msg.X
					m.selection.endY = msg.Y
					if !m.selection.empty() {
						text := m.extractSelectedText()
						go copyToClipboard(text)
						m.copyToast = true
						id := m.id
						return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
							return ClearToastMsg{ID: id}
						})
					} else {
						m.selection = textSelection{}
					}
					return m, nil
				}
			}
		}

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
			orig := m.emu.CellAt(cx, cy)
			var cell *uv.Cell
			switch m.cursorState.style {
		case vt.CursorUnderline:
			cell = &uv.Cell{Content: "_"}
			fg, bg := orig.Style.Fg, orig.Style.Bg
			if fg == nil {
				fg = m.emu.ForegroundColor()
			}
			if bg == nil {
				bg = m.emu.BackgroundColor()
			}
			saved = orig.Clone()
			cell.Style.Fg, cell.Style.Bg = bg, fg
		case vt.CursorBar:
			cell = &uv.Cell{Content: "▏"}
			fg, bg := orig.Style.Fg, orig.Style.Bg
			if fg == nil {
				fg = m.emu.ForegroundColor()
			}
			if bg == nil {
				bg = m.emu.BackgroundColor()
			}
			saved = orig.Clone()
			cell.Style.Fg, cell.Style.Bg = bg, fg
		default: // CursorBlock
			fg, bg := orig.Style.Fg, orig.Style.Bg
			if fg == nil {
				fg = m.emu.ForegroundColor()
			}
			if bg == nil {
				bg = m.emu.BackgroundColor()
			}
			cell = orig.Clone()
			saved = orig.Clone()
			cell.Style.Fg, cell.Style.Bg = bg, fg
			}
			m.emu.SetCell(cx, cy, cell)
		}
	}

	restoreSelection := m.applySelectionHighlight()

	var result string
	if m.emu.IsAltScreen() || (m.scrollOffset == 0 && m.emu.ScrollbackLen() == 0) {
		result = strings.TrimRight(m.emu.Render(), "\n")
	} else if m.scrollOffset == 0 {
		result = m.overlayScrollbar(strings.TrimRight(m.emu.Render(), "\n"))
	} else {
		result = m.renderScrolled()
	}

	restoreSelection()
	if saved != nil {
		m.emu.SetCell(cx, cy, saved)
	}

	if m.copyToast && m.width > 0 && m.height > 0 {
		result = m.overlayToast(result, "Copied!")
	}

	return result
}

func (m TermViewModel) renderScrolled() string {
	if m.height <= 0 {
		return ""
	}
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

func (m TermViewModel) overlayToast(content, msg string) string {
	toastStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.ANSIColor(0)).
		Background(lipgloss.ANSIColor(2)).
		Padding(0, 1)
	toast := toastStyle.Render(msg)
	toastW := lipgloss.Width(toast)

	lines := strings.Split(content, "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	row := 0
	col := m.width - toastW
	if col < 0 {
		col = 0
	}

	if row < len(lines) {
		line := lines[row]
		prefix := ansi.Truncate(line, col, "")
		lines[row] = prefix + toast
	}

	return strings.Join(lines[:min(len(lines), m.height)], "\n")
}

func ResizePTY(f *os.File, rows, cols uint16) error {
	return pty.Setsize(f, &pty.Winsize{Rows: rows, Cols: cols})
}

// respondPaletteQueries scans data for OSC 4 palette queries
// ("\x1b]4;<index>;?") from the child and writes the corresponding color back
// to the PTY. The VT emulator has no OSC 4 handler, so without this a child
// that builds its theme from the terminal palette (e.g. OpenCode) gets no
// answer and falls back to its default theme. Indices 0-15 use the host
// palette; higher indices use the standard xterm 256-color values.
func respondPaletteQueries(ptyFile *os.File, data []byte) {
	const prefix = "\x1b]4;"
	rest := data
	for {
		i := bytes.Index(rest, []byte(prefix))
		if i < 0 {
			return
		}
		rest = rest[i+len(prefix):]

		// Body runs until the OSC terminator (BEL or ST).
		end := len(rest)
		for j := 0; j < len(rest); j++ {
			if rest[j] == 0x07 {
				end = j
				break
			}
			if rest[j] == 0x1b && j+1 < len(rest) && rest[j+1] == '\\' {
				end = j
				break
			}
		}
		body := rest[:end]
		rest = rest[end:]

		// Body is "index;spec[;index;spec...]"; respond to each "?" spec.
		parts := bytes.Split(body, []byte{';'})
		for k := 0; k+1 < len(parts); k += 2 {
			if string(parts[k+1]) != "?" {
				continue
			}
			idx, err := strconv.Atoi(string(parts[k]))
			if err != nil || idx < 0 || idx > 255 {
				continue
			}
			c := paletteColorFor(idx)
			if c == nil {
				continue
			}
			r, g, b, _ := c.RGBA()
			reply := fmt.Sprintf("\x1b]4;%d;rgb:%04x/%04x/%04x\x07", idx, r, g, b)
			if _, err := ptyFile.Write([]byte(reply)); err != nil {
				return
			}
		}
	}
}

// paletteColorFor returns the color to report for an OSC 4 palette index.
func paletteColorFor(idx int) color.Color {
	colors := getHostColors()
	if idx >= 0 && idx < 16 {
		if c := colors.palette[idx]; c != nil {
			return c
		}
	}
	return ansi.IndexedColor(uint8(idx)) //nolint:gosec
}

func applyHostColors(emu *vt.SafeEmulator, colors hostTermColors) {
	if colors.fg != nil {
		emu.SetDefaultForegroundColor(colors.fg)
		emu.SetForegroundColor(colors.fg)
	}
	if colors.bg != nil {
		emu.SetDefaultBackgroundColor(colors.bg)
		emu.SetBackgroundColor(colors.bg)
	}
	for i, c := range colors.palette {
		if c != nil {
			emu.SetIndexedColor(i, c)
		}
	}
}

// UpdateHostColors pushes the current host terminal colors to the emulator.
func (m TermViewModel) UpdateHostColors() TermViewModel {
	applyHostColors(m.emu, getHostColors())
	return m
}

// ColorSchemeChangedMsg is sent when the system color scheme changes.
type ColorSchemeChangedMsg struct{}

func isMouseTrackingMode(mode ansi.Mode) bool {
	dm, ok := mode.(ansi.DECMode)
	if !ok {
		return false
	}
	switch dm {
	case ansi.ModeMouseX10, ansi.ModeMouseNormal, ansi.ModeMouseHighlight,
		ansi.ModeMouseButtonEvent, ansi.ModeMouseAnyEvent:
		return true
	}
	return false
}

func (m TermViewModel) viewStart() int {
	sb := m.emu.Scrollback()
	sbLen := 0
	if sb != nil {
		sbLen = sb.Len()
	}
	totalLines := sbLen + m.height
	vs := totalLines - m.height - m.scrollOffset
	if vs < 0 {
		vs = 0
	}
	return vs
}

func (m TermViewModel) applySelectionHighlight() func() {
	sel := m.selection
	if sel.empty() && !sel.active {
		return func() {}
	}

	sx, sy, ex, ey := sel.normalized()
	if sy >= m.height || ey < 0 {
		return func() {}
	}
	if sy < 0 {
		sy = 0
		sx = 0
	}
	if ey >= m.height {
		ey = m.height - 1
		ex = m.width - 1
	}
	if sx < 0 {
		sx = 0
	}
	if ex >= m.width {
		ex = m.width - 1
	}

	sb := m.emu.Scrollback()
	sbLen := 0
	if sb != nil {
		sbLen = sb.Len()
	}
	vs := m.viewStart()
	fg := m.emu.ForegroundColor()
	bg := m.emu.BackgroundColor()

	type savedEntry struct {
		cell    uv.Cell
		x       int
		scrY    int
		sbLineY int
		isSB    bool
	}
	var saved []savedEntry

	for vy := sy; vy <= ey; vy++ {
		colStart := 0
		colEnd := m.width - 1
		if vy == sy {
			colStart = sx
		}
		if vy == ey {
			colEnd = ex
		}

		lineIdx := vs + vy
		isSB := lineIdx < sbLen

		for x := colStart; x <= colEnd; x++ {
			var orig *uv.Cell
			var scrY int

			if isSB {
				line := sb.Line(lineIdx)
				if x < len(line) {
					orig = line.At(x)
				}
			} else {
				scrY = lineIdx - sbLen
				orig = m.emu.CellAt(x, scrY)
			}

			if orig == nil || orig.IsZero() {
				continue
			}

			saved = append(saved, savedEntry{
				cell: *orig, x: x, scrY: scrY, sbLineY: lineIdx, isSB: isSB,
			})

			cellFg, cellBg := orig.Style.Fg, orig.Style.Bg
			if cellFg == nil {
				cellFg = fg
			}
			if cellBg == nil {
				cellBg = bg
			}

			if isSB {
				line := sb.Line(lineIdx)
				if x < len(line) {
					line[x].Style.Fg = cellBg
					line[x].Style.Bg = cellFg
				}
			} else {
				inverted := orig.Clone()
				inverted.Style.Fg = cellBg
				inverted.Style.Bg = cellFg
				m.emu.SetCell(x, scrY, inverted)
			}
		}
	}

	return func() {
		for _, e := range saved {
			if e.isSB {
				line := sb.Line(e.sbLineY)
				if e.x < len(line) {
					line[e.x] = e.cell
				}
			} else {
				restored := e.cell
				m.emu.SetCell(e.x, e.scrY, &restored)
			}
		}
	}
}

func (m TermViewModel) extractSelectedText() string {
	sel := m.selection
	if sel.empty() {
		return ""
	}

	sx, sy, ex, ey := sel.normalized()
	if sy < 0 {
		sy = 0
		sx = 0
	}
	if ey >= m.height {
		ey = m.height - 1
		ex = m.width - 1
	}
	if sx < 0 {
		sx = 0
	}
	if ex >= m.width {
		ex = m.width - 1
	}

	sb := m.emu.Scrollback()
	sbLen := 0
	if sb != nil {
		sbLen = sb.Len()
	}
	vs := m.viewStart()

	var lines []string
	for vy := sy; vy <= ey; vy++ {
		colStart := 0
		colEnd := m.width - 1
		if vy == sy {
			colStart = sx
		}
		if vy == ey {
			colEnd = ex
		}

		lineIdx := vs + vy
		isSB := lineIdx < sbLen

		var buf strings.Builder
		for x := colStart; x <= colEnd; x++ {
			var cell *uv.Cell
			if isSB {
				cell = sb.CellAt(x, lineIdx)
			} else {
				cell = m.emu.CellAt(x, lineIdx-sbLen)
			}
			if cell == nil {
				buf.WriteByte(' ')
			} else if cell.IsZero() {
				continue
			} else if cell.Equal(&uv.EmptyCell) {
				buf.WriteByte(' ')
			} else {
				buf.WriteString(cell.Content)
			}
		}
		lines = append(lines, strings.TrimRight(buf.String(), " "))
	}

	return strings.Join(lines, "\n")
}

func copyToClipboard(text string) {
	if path, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
		return
	}
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
		return
	}
	if path, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command(path, "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
		return
	}
	if path, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
	}
}
