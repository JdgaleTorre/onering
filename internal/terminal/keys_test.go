package terminal

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestKeyMsgToUV_SpecialKeys(t *testing.T) {
	tests := []struct {
		name    string
		keyType tea.KeyType
		code    rune
		mod     uv.KeyMod
	}{
		{"Enter", tea.KeyEnter, uv.KeyEnter, 0},
		{"Tab", tea.KeyTab, uv.KeyTab, 0},
		{"Backspace", tea.KeyBackspace, uv.KeyBackspace, 0},
		{"Escape", tea.KeyEsc, uv.KeyEscape, 0},
		{"Up", tea.KeyUp, uv.KeyUp, 0},
		{"Down", tea.KeyDown, uv.KeyDown, 0},
		{"Left", tea.KeyLeft, uv.KeyLeft, 0},
		{"Right", tea.KeyRight, uv.KeyRight, 0},
		{"Home", tea.KeyHome, uv.KeyHome, 0},
		{"End", tea.KeyEnd, uv.KeyEnd, 0},
		{"PgUp", tea.KeyPgUp, uv.KeyPgUp, 0},
		{"PgDown", tea.KeyPgDown, uv.KeyPgDown, 0},
		{"Delete", tea.KeyDelete, uv.KeyDelete, 0},
		{"Insert", tea.KeyInsert, uv.KeyInsert, 0},
		{"ShiftTab", tea.KeyShiftTab, uv.KeyTab, uv.ModShift},
		{"CtrlUp", tea.KeyCtrlUp, uv.KeyUp, uv.ModCtrl},
		{"CtrlDown", tea.KeyCtrlDown, uv.KeyDown, uv.ModCtrl},
		{"CtrlRight", tea.KeyCtrlRight, uv.KeyRight, uv.ModCtrl},
		{"CtrlLeft", tea.KeyCtrlLeft, uv.KeyLeft, uv.ModCtrl},
		{"CtrlShiftLeft", tea.KeyCtrlShiftLeft, uv.KeyLeft, uv.ModCtrl | uv.ModShift},
		{"CtrlShiftRight", tea.KeyCtrlShiftRight, uv.KeyRight, uv.ModCtrl | uv.ModShift},
		{"ShiftUp", tea.KeyShiftUp, uv.KeyUp, uv.ModShift},
		{"ShiftDown", tea.KeyShiftDown, uv.KeyDown, uv.ModShift},
		{"F1", tea.KeyF1, uv.KeyF1, 0},
		{"F12", tea.KeyF12, uv.KeyF12, 0},
		{"F20", tea.KeyF20, uv.KeyF20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tt.keyType}
			got, ok := keyMsgToUV(msg)
			if !ok {
				t.Fatal("expected ok=true")
			}
			key := uv.Key(got)
			if key.Code != tt.code {
				t.Errorf("Code = %d, want %d", key.Code, tt.code)
			}
			if key.Mod != tt.mod {
				t.Errorf("Mod = %d, want %d", key.Mod, tt.mod)
			}
		})
	}
}

func TestKeyMsgToUV_Space(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeySpace}
	got, ok := keyMsgToUV(msg)
	if !ok {
		t.Fatal("expected ok=true")
	}
	key := uv.Key(got)
	if key.Code != uv.KeySpace {
		t.Errorf("Code = %d, want KeySpace", key.Code)
	}
	if key.Text != " " {
		t.Errorf("Text = %q, want %q", key.Text, " ")
	}
}

func TestKeyMsgToUV_Runes(t *testing.T) {
	t.Run("single rune", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
		got, ok := keyMsgToUV(msg)
		if !ok {
			t.Fatal("expected ok=true")
		}
		key := uv.Key(got)
		if key.Code != 'a' {
			t.Errorf("Code = %d, want 'a'", key.Code)
		}
		if key.Text != "a" {
			t.Errorf("Text = %q, want %q", key.Text, "a")
		}
	})

	t.Run("empty runes returns false", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
		_, ok := keyMsgToUV(msg)
		if ok {
			t.Error("expected ok=false for empty runes")
		}
	})
}

func TestKeyMsgToUV_CtrlKeys(t *testing.T) {
	t.Run("Ctrl+A through Ctrl+Z", func(t *testing.T) {
		// Ctrl+I aliases to KeyTab, Ctrl+M aliases to KeyEnter in bubbletea,
		// so they match the specialKeys map instead of the Ctrl+letter handler.
		aliased := map[tea.KeyType]bool{
			tea.KeyTab:   true, // Ctrl+I
			tea.KeyEnter: true, // Ctrl+M
		}
		for i := tea.KeyCtrlA; i <= tea.KeyCtrlZ; i++ {
			if aliased[i] {
				continue
			}
			msg := tea.KeyMsg{Type: i}
			got, ok := keyMsgToUV(msg)
			if !ok {
				t.Errorf("expected ok=true for KeyType %d", i)
				continue
			}
			key := uv.Key(got)
			wantCode := rune('a' + i - tea.KeyCtrlA)
			if key.Code != wantCode {
				t.Errorf("KeyType %d: Code = %d (%c), want %d (%c)", i, key.Code, key.Code, wantCode, wantCode)
			}
			if key.Mod&uv.ModCtrl == 0 {
				t.Errorf("KeyType %d: expected ModCtrl", i)
			}
		}
	})

	ctrlSpecial := []struct {
		name    string
		keyType tea.KeyType
		code    rune
	}{
		{"Ctrl+@", tea.KeyCtrlAt, '@'},
		{"Ctrl+Backslash", tea.KeyCtrlBackslash, '\\'},
		{"Ctrl+CloseBracket", tea.KeyCtrlCloseBracket, ']'},
		{"Ctrl+Caret", tea.KeyCtrlCaret, '^'},
		{"Ctrl+Underscore", tea.KeyCtrlUnderscore, '_'},
	}
	for _, tt := range ctrlSpecial {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tt.keyType}
			got, ok := keyMsgToUV(msg)
			if !ok {
				t.Fatal("expected ok=true")
			}
			key := uv.Key(got)
			if key.Code != tt.code {
				t.Errorf("Code = %d (%c), want %d (%c)", key.Code, key.Code, tt.code, tt.code)
			}
			if key.Mod&uv.ModCtrl == 0 {
				t.Errorf("expected ModCtrl")
			}
		})
	}
}

func TestKeyMsgToUV_AltModifier(t *testing.T) {
	t.Run("Alt+Enter", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter, Alt: true}
		got, ok := keyMsgToUV(msg)
		if !ok {
			t.Fatal("expected ok=true")
		}
		key := uv.Key(got)
		if key.Code != uv.KeyEnter {
			t.Errorf("Code = %d, want KeyEnter", key.Code)
		}
		if key.Mod&uv.ModAlt == 0 {
			t.Error("expected ModAlt")
		}
		if key.Text != "" {
			t.Errorf("Text = %q, want empty (cleared by Alt)", key.Text)
		}
	})

	t.Run("Alt+rune clears text", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true}
		got, ok := keyMsgToUV(msg)
		if !ok {
			t.Fatal("expected ok=true")
		}
		key := uv.Key(got)
		if key.Code != 'x' {
			t.Errorf("Code = %d, want 'x'", key.Code)
		}
		if key.Mod&uv.ModAlt == 0 {
			t.Error("expected ModAlt")
		}
		if key.Text != "" {
			t.Errorf("Text = %q, want empty (cleared by Alt)", key.Text)
		}
	})
}

func TestKeyMsgToUV_UnknownKey(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyType(-999)}
	_, ok := keyMsgToUV(msg)
	if ok {
		t.Error("expected ok=false for unknown key type")
	}
}
