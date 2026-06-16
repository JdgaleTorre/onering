//go:build !unix

package terminal

import (
	"image/color"
	"os"
)

// queryPalette is a no-op on platforms without a unix tty. The emulator falls
// back to the standard xterm palette.
func queryPalette(_ *os.File, _ *[16]color.Color) {}
