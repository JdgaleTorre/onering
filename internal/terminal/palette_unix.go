//go:build unix

package terminal

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

// paletteReadTimeout bounds how long we wait for each byte of an OSC 4
// response before giving up. Terminals that support palette queries answer
// almost immediately; this only guards against terminals that silently ignore
// the query.
const paletteReadTimeout = 200 * time.Millisecond

// queryPalette asks the terminal for its 16 ANSI palette colors (OSC 4) and
// stores any answers in dst. The tty must be a real terminal. It puts the tty
// into raw mode for the duration so the responses are not echoed or
// line-buffered, then restores the previous state.
func queryPalette(tty *os.File, dst *[16]color.Color) {
	fd := tty.Fd()
	state, err := term.MakeRaw(fd)
	if err != nil {
		return
	}
	defer term.Restore(fd, state) //nolint:errcheck

	// Send all 16 palette queries, then a cursor-position request as a
	// sentinel: terminals answer requests in order, so once the CSI R reply
	// arrives we know every palette answer that is coming has arrived.
	for i := range 16 {
		fmt.Fprintf(tty, "\x1b]4;%d;?\x07", i) //nolint:errcheck
	}
	fmt.Fprint(tty, "\x1b[6n") //nolint:errcheck

	for {
		resp, isOSC, err := readResponse(int(fd))
		if err != nil {
			return
		}
		if !isOSC {
			// The cursor-position reply: all palette answers are in.
			return
		}
		if idx, c, ok := parsePaletteResponse(resp); ok {
			dst[idx] = c
		}
	}
}

// readResponse reads a single OSC ("\x1b]...BEL/ST") or CSI cursor-position
// ("\x1b[...R") reply from fd, byte by byte, with a per-byte timeout.
func readResponse(fd int) (resp string, isOSC bool, err error) {
	readByte := func() (byte, error) {
		var rfds unix.FdSet
		rfds.Set(fd)
		tv := unix.NsecToTimeval(int64(paletteReadTimeout))
		for {
			n, serr := unix.Select(fd+1, &rfds, nil, nil, &tv)
			if serr == unix.EINTR {
				continue
			}
			if serr != nil {
				return 0, serr
			}
			if n == 0 {
				return 0, fmt.Errorf("timeout")
			}
			break
		}
		var b [1]byte
		if _, rerr := unix.Read(fd, b[:]); rerr != nil {
			return 0, rerr
		}
		return b[0], nil
	}

	// Skip to the ESC that starts a response.
	b, err := readByte()
	if err != nil {
		return "", false, err
	}
	for b != 0x1b {
		if b, err = readByte(); err != nil {
			return "", false, err
		}
	}

	tpe, err := readByte()
	if err != nil {
		return "", false, err
	}
	switch tpe {
	case ']':
		isOSC = true
	case '[':
		isOSC = false
	default:
		return "", false, fmt.Errorf("unexpected response type %q", tpe)
	}

	resp = "\x1b" + string(tpe)
	for {
		b, err := readByte()
		if err != nil {
			return "", false, err
		}
		resp += string(b)
		if isOSC {
			if b == 0x07 || (b == '\\' && len(resp) >= 2 && resp[len(resp)-2] == 0x1b) {
				return resp, true, nil
			}
		} else if b == 'R' {
			return resp, false, nil
		}
		if len(resp) > 64 {
			return "", false, fmt.Errorf("response too long")
		}
	}
}

// parsePaletteResponse parses "\x1b]4;<index>;rgb:RRRR/GGGG/BBBB<terminator>".
func parsePaletteResponse(resp string) (index int, c color.Color, ok bool) {
	// Strip the "\x1b]4;" prefix and the terminator.
	body := resp
	if len(body) < 5 || body[:4] != "\x1b]4;" {
		return 0, nil, false
	}
	body = body[4:]
	body = trimTerminator(body)

	semi := -1
	for i := 0; i < len(body); i++ {
		if body[i] == ';' {
			semi = i
			break
		}
	}
	if semi <= 0 {
		return 0, nil, false
	}
	idx, err := strconv.Atoi(body[:semi])
	if err != nil || idx < 0 || idx > 15 {
		return 0, nil, false
	}
	parsed := ansi.XParseColor(body[semi+1:])
	if parsed == nil {
		return 0, nil, false
	}
	return idx, parsed, true
}

func trimTerminator(s string) string {
	if n := len(s); n > 0 && s[n-1] == 0x07 { // BEL
		return s[:n-1]
	}
	if n := len(s); n >= 2 && s[n-2] == 0x1b && s[n-1] == '\\' { // ST
		return s[:n-2]
	}
	return s
}
