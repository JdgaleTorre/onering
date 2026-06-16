package terminal

import (
	"image/color"
	"os"
	"sync"

	"github.com/muesli/termenv"
)

var (
	hostColors     hostTermColors
	hostColorsOnce sync.Once
)

type hostTermColors struct {
	fg, bg  color.Color
	palette [16]color.Color
}

// DetectHostColors queries the real terminal for its foreground, background,
// and 16 ANSI palette colors so embedded child terminals can be made to match
// the host. Must be called before Bubbletea takes over stdin.
func DetectHostColors() {
	hostColorsOnce.Do(detectColors)
}

func detectColors() {
	// Try /dev/tty first (works even if stdout is redirected); fall back to
	// stdout for normal terminal sessions.
	if tryDevTTY() {
		return
	}
	tryStdout()
}

func tryDevTTY() bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	defer tty.Close()
	ok := queryTermenv(termenv.NewOutput(tty, termenv.WithTTY(true)))
	// termenv restores the tty to cooked mode before returning, so it is safe
	// to run our own raw-mode palette query on the same handle afterwards.
	queryPalette(tty, &hostColors.palette)
	return ok
}

func tryStdout() bool {
	return queryTermenv(termenv.DefaultOutput())
}

func queryTermenv(out *termenv.Output) bool {
	got := false
	if fgRaw := out.ForegroundColor(); fgRaw != nil {
		if _, ok := fgRaw.(termenv.NoColor); !ok {
			hostColors.fg = termenv.ConvertToRGB(fgRaw)
			got = true
		}
	}
	if bgRaw := out.BackgroundColor(); bgRaw != nil {
		if _, ok := bgRaw.(termenv.NoColor); !ok {
			hostColors.bg = termenv.ConvertToRGB(bgRaw)
			got = true
		}
	}
	return got
}
