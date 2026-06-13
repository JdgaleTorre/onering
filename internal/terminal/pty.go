package terminal

import (
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

type PTYHandle struct {
	F   *os.File
	Cmd *exec.Cmd
}

func StartPTY(cmd *exec.Cmd) (*PTYHandle, error) {
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &PTYHandle{F: f, Cmd: cmd}, nil
}

func (h *PTYHandle) Write(p []byte) (int, error) {
	return h.F.Write(p)
}

func (h *PTYHandle) Read(p []byte) (int, error) {
	return h.F.Read(p)
}

func (h *PTYHandle) Resize(rows, cols uint16) error {
	return pty.Setsize(h.F, &pty.Winsize{Rows: rows, Cols: cols})
}

func (h *PTYHandle) GracefulClose(input string) error {
	if h.Cmd == nil || h.Cmd.Process == nil {
		return h.F.Close()
	}
	h.F.Write([]byte(input + "\n"))
	done := make(chan struct{})
	go func() {
		h.Cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		if h.Cmd.Process != nil {
			h.Cmd.Process.Kill()
		}
	}
	return h.F.Close()
}

func (h *PTYHandle) Close() error {
	if h.Cmd != nil && h.Cmd.Process != nil {
		h.Cmd.Process.Kill()
	}
	return h.F.Close()
}
