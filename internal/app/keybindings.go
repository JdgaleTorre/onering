package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	FocusLeft  key.Binding
	FocusRight key.Binding
	Up         key.Binding
	Down       key.Binding
	Insert     key.Binding
	Escape     key.Binding
	Enter      key.Binding
	NewSession key.Binding
	Delete     key.Binding
	SwitchAgent key.Binding
	Editor     key.Binding
	Git        key.Binding
	Fullscreen key.Binding
	Tab        key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	FocusLeft:  key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "focus sidebar")),
	FocusRight: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "focus main")),
	Up:         key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "up")),
	Down:       key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "down")),
	Insert:     key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "insert mode")),
	Escape:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "navigation mode")),
	Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send/select")),
	NewSession: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new session")),
	Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete session")),
	SwitchAgent: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "switch agent")),
	Editor:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "open editor")),
	Git:        key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "open lazygit")),
	Fullscreen: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fullscreen")),
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle focus")),
	PageUp:     key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "page up")),
	PageDown:   key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "page down")),
}

func (k KeyMap) NavigationBindings() []key.Binding {
	return []key.Binding{
		k.Quit, k.Help, k.FocusLeft, k.FocusRight,
		k.Up, k.Down, k.Insert, k.NewSession, k.Delete,
		k.SwitchAgent, k.Editor, k.Git, k.Fullscreen,
		k.Tab, k.PageUp, k.PageDown,
	}
}

func (k KeyMap) InsertBindings() []key.Binding {
	return []key.Binding{
		k.Escape, k.Enter,
	}
}
