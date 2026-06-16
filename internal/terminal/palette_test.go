package terminal

import (
	"image/color"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// rgb is a convenient opaque color whose RGBA() expands each 8-bit channel into
// the doubled 16-bit form terminals report (e.g. 0xc3 -> 0xc3c3).
func rgb(r, g, b uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

func colorsEqual(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

// withPalette temporarily installs a host palette for the duration of a test.
func withPalette(t *testing.T, p [16]color.Color) {
	t.Helper()
	prev := hostColors.palette
	hostColors.palette = p
	t.Cleanup(func() { hostColors.palette = prev })
}

func TestRespondPaletteQueries(t *testing.T) {
	var p [16]color.Color
	p[0] = rgb(0x09, 0x06, 0x18)
	p[1] = rgb(0xc3, 0x40, 0x43)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "single index",
			in:   "\x1b]4;1;?\x07",
			want: "\x1b]4;1;rgb:c3c3/4040/4343\x07",
		},
		{
			name: "batched indices in one OSC",
			in:   "\x1b]4;0;?;1;?\x07",
			want: "\x1b]4;0;rgb:0909/0606/1818\x07\x1b]4;1;rgb:c3c3/4040/4343\x07",
		},
		{
			name: "separate OSC sequences",
			in:   "\x1b]4;0;?\x07\x1b]4;1;?\x07",
			want: "\x1b]4;0;rgb:0909/0606/1818\x07\x1b]4;1;rgb:c3c3/4040/4343\x07",
		},
		{
			name: "ST terminator",
			in:   "\x1b]4;1;?\x1b\\",
			want: "\x1b]4;1;rgb:c3c3/4040/4343\x07",
		},
		{
			name: "set (not query) is ignored",
			in:   "\x1b]4;1;rgb:1111/2222/3333\x07",
			want: "",
		},
		{
			name: "no OSC 4 at all",
			in:   "hello world\x1b[0m",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPalette(t, p)
			got := captureReplies(t, []byte(tt.in))
			if got != tt.want {
				t.Errorf("reply = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRespondPaletteQueries_HighIndexUsesXtermDefault(t *testing.T) {
	withPalette(t, [16]color.Color{}) // empty host palette
	got := captureReplies(t, []byte("\x1b]4;200;?\x07"))

	r, g, b, _ := ansi.IndexedColor(200).RGBA()
	want := sprintfReply(200, r, g, b)
	if got != want {
		t.Errorf("reply = %q, want %q", got, want)
	}
}

func TestPaletteColorFor(t *testing.T) {
	var p [16]color.Color
	p[2] = rgb(0x76, 0x94, 0x6a)
	withPalette(t, p)

	if c := paletteColorFor(2); !colorsEqual(c, p[2]) {
		t.Errorf("index 2: got %v, want host palette color", c)
	}
	// Index 5 has no host color set: fall back to the xterm default.
	if c := paletteColorFor(5); !colorsEqual(c, ansi.IndexedColor(5)) {
		t.Errorf("index 5: got %v, want xterm default", c)
	}
	// Index above the 16-color range always uses the xterm default.
	if c := paletteColorFor(200); !colorsEqual(c, ansi.IndexedColor(200)) {
		t.Errorf("index 200: got %v, want xterm default", c)
	}
}

// captureReplies runs respondPaletteQueries against an in-memory pipe and
// returns everything it wrote.
func captureReplies(t *testing.T, in []byte) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	respondPaletteQueries(w, in)
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	r.Close()
	return string(out)
}

func sprintfReply(idx int, r, g, b uint32) string {
	const hex = "0123456789abcdef"
	enc := func(v uint32) string {
		return string([]byte{
			hex[(v>>12)&0xf], hex[(v>>8)&0xf], hex[(v>>4)&0xf], hex[v&0xf],
		})
	}
	return "\x1b]4;" + itoa(idx) + ";rgb:" + enc(r) + "/" + enc(g) + "/" + enc(b) + "\x07"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
