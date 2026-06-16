//go:build unix

package terminal

import "testing"

func TestParsePaletteResponse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantIdx int
		wantR   uint32
		wantG   uint32
		wantB   uint32
		wantOK  bool
	}{
		{
			name:    "BEL terminated",
			in:      "\x1b]4;5;rgb:9595/7f7f/b8b8\x07",
			wantIdx: 5, wantR: 0x9595, wantG: 0x7f7f, wantB: 0xb8b8, wantOK: true,
		},
		{
			name:    "ST terminated",
			in:      "\x1b]4;1;rgb:c3c3/4040/4343\x1b\\",
			wantIdx: 1, wantR: 0xc3c3, wantG: 0x4040, wantB: 0x4343, wantOK: true,
		},
		{
			name:    "double digit index",
			in:      "\x1b]4;15;rgb:dcdc/d7d7/baba\x07",
			wantIdx: 15, wantR: 0xdcdc, wantG: 0xd7d7, wantB: 0xbaba, wantOK: true,
		},
		{name: "wrong prefix", in: "\x1b]10;rgb:1111/1111/1111\x07", wantOK: false},
		{name: "index out of range", in: "\x1b]4;99;rgb:1111/1111/1111\x07", wantOK: false},
		{name: "missing color", in: "\x1b]4;3\x07", wantOK: false},
		{name: "unparseable color", in: "\x1b]4;3;notacolor\x07", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, c, ok := parsePaletteResponse(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if idx != tt.wantIdx {
				t.Errorf("idx = %d, want %d", idx, tt.wantIdx)
			}
			r, g, b, _ := c.RGBA()
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("color = %04x/%04x/%04x, want %04x/%04x/%04x",
					r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestTrimTerminator(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"abc\x07", "abc"},
		{"abc\x1b\\", "abc"},
		{"abc", "abc"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := trimTerminator(tt.in); got != tt.want {
			t.Errorf("trimTerminator(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
