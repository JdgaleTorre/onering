package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
)

var specialKeys = map[tea.KeyType]uv.Key{
	tea.KeyEnter:          {Code: uv.KeyEnter},
	tea.KeyTab:            {Code: uv.KeyTab},
	tea.KeyBackspace:      {Code: uv.KeyBackspace},
	tea.KeyEsc:            {Code: uv.KeyEscape},
	tea.KeySpace:          {Code: uv.KeySpace, Text: " "},
	tea.KeyUp:             {Code: uv.KeyUp},
	tea.KeyDown:           {Code: uv.KeyDown},
	tea.KeyRight:          {Code: uv.KeyRight},
	tea.KeyLeft:           {Code: uv.KeyLeft},
	tea.KeyShiftTab:       {Code: uv.KeyTab, Mod: uv.ModShift},
	tea.KeyHome:           {Code: uv.KeyHome},
	tea.KeyEnd:            {Code: uv.KeyEnd},
	tea.KeyPgUp:           {Code: uv.KeyPgUp},
	tea.KeyPgDown:         {Code: uv.KeyPgDown},
	tea.KeyCtrlPgUp:       {Code: uv.KeyPgUp, Mod: uv.ModCtrl},
	tea.KeyCtrlPgDown:     {Code: uv.KeyPgDown, Mod: uv.ModCtrl},
	tea.KeyDelete:         {Code: uv.KeyDelete},
	tea.KeyInsert:         {Code: uv.KeyInsert},
	tea.KeyCtrlUp:         {Code: uv.KeyUp, Mod: uv.ModCtrl},
	tea.KeyCtrlDown:       {Code: uv.KeyDown, Mod: uv.ModCtrl},
	tea.KeyCtrlRight:      {Code: uv.KeyRight, Mod: uv.ModCtrl},
	tea.KeyCtrlLeft:       {Code: uv.KeyLeft, Mod: uv.ModCtrl},
	tea.KeyCtrlHome:       {Code: uv.KeyHome, Mod: uv.ModCtrl},
	tea.KeyCtrlEnd:        {Code: uv.KeyEnd, Mod: uv.ModCtrl},
	tea.KeyShiftUp:        {Code: uv.KeyUp, Mod: uv.ModShift},
	tea.KeyShiftDown:      {Code: uv.KeyDown, Mod: uv.ModShift},
	tea.KeyShiftRight:     {Code: uv.KeyRight, Mod: uv.ModShift},
	tea.KeyShiftLeft:      {Code: uv.KeyLeft, Mod: uv.ModShift},
	tea.KeyShiftHome:      {Code: uv.KeyHome, Mod: uv.ModShift},
	tea.KeyShiftEnd:       {Code: uv.KeyEnd, Mod: uv.ModShift},
	tea.KeyCtrlShiftUp:    {Code: uv.KeyUp, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyCtrlShiftDown:  {Code: uv.KeyDown, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyCtrlShiftLeft:  {Code: uv.KeyLeft, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyCtrlShiftRight: {Code: uv.KeyRight, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyCtrlShiftHome:  {Code: uv.KeyHome, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyCtrlShiftEnd:   {Code: uv.KeyEnd, Mod: uv.ModCtrl | uv.ModShift},
	tea.KeyF1:             {Code: uv.KeyF1},
	tea.KeyF2:             {Code: uv.KeyF2},
	tea.KeyF3:             {Code: uv.KeyF3},
	tea.KeyF4:             {Code: uv.KeyF4},
	tea.KeyF5:             {Code: uv.KeyF5},
	tea.KeyF6:             {Code: uv.KeyF6},
	tea.KeyF7:             {Code: uv.KeyF7},
	tea.KeyF8:             {Code: uv.KeyF8},
	tea.KeyF9:             {Code: uv.KeyF9},
	tea.KeyF10:            {Code: uv.KeyF10},
	tea.KeyF11:            {Code: uv.KeyF11},
	tea.KeyF12:            {Code: uv.KeyF12},
	tea.KeyF13:            {Code: uv.KeyF13},
	tea.KeyF14:            {Code: uv.KeyF14},
	tea.KeyF15:            {Code: uv.KeyF15},
	tea.KeyF16:            {Code: uv.KeyF16},
	tea.KeyF17:            {Code: uv.KeyF17},
	tea.KeyF18:            {Code: uv.KeyF18},
	tea.KeyF19:            {Code: uv.KeyF19},
	tea.KeyF20:            {Code: uv.KeyF20},
}

// keyMsgToUV translates a bubbletea v1 key message into the key event
// type the vt emulator understands, so it can encode the key according
// to the modes the child program enabled (application cursor keys, ...).
func keyMsgToUV(msg tea.KeyMsg) (uv.KeyPressEvent, bool) {
	k, ok := specialKeys[msg.Type]
	if !ok {
		switch t := msg.Type; {
		case t == tea.KeyRunes:
			if len(msg.Runes) == 0 {
				return uv.KeyPressEvent{}, false
			}
			k = uv.Key{Code: msg.Runes[0], Text: string(msg.Runes)}
		case t >= tea.KeyCtrlA && t <= tea.KeyCtrlZ:
			k = uv.Key{Code: rune('a' + t - tea.KeyCtrlA), Mod: uv.ModCtrl}
		case t == tea.KeyCtrlAt:
			k = uv.Key{Code: '@', Mod: uv.ModCtrl}
		case t == tea.KeyCtrlBackslash:
			k = uv.Key{Code: '\\', Mod: uv.ModCtrl}
		case t == tea.KeyCtrlCloseBracket:
			k = uv.Key{Code: ']', Mod: uv.ModCtrl}
		case t == tea.KeyCtrlCaret:
			k = uv.Key{Code: '^', Mod: uv.ModCtrl}
		case t == tea.KeyCtrlUnderscore:
			k = uv.Key{Code: '_', Mod: uv.ModCtrl}
		default:
			return uv.KeyPressEvent{}, false
		}
	}
	if msg.Alt {
		k.Mod |= uv.ModAlt
		k.Text = ""
	}
	return uv.KeyPressEvent(k), true
}
