package terminal

import (
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

const gracefulCloseTimeout = 3 * time.Second

type PTYHandle struct {
	F   *os.File
	Cmd *exec.Cmd
}

func StartPTY(cmd *exec.Cmd) (*PTYHandle, error) {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	ensureEnv(cmd, "TERM", "xterm-256color")
	ensureEnv(cmd, "COLORTERM", "truecolor")
	forwardEnv(cmd, "TERM_PROGRAM")
	forwardEnv(cmd, "TERM_PROGRAM_VERSION")

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &PTYHandle{F: f, Cmd: cmd}, nil
}

func ensureEnv(cmd *exec.Cmd, key, fallback string) {
	prefix := key + "="
	for _, e := range cmd.Env {
		if len(e) > len(prefix) && e[:len(prefix)] == prefix {
			return
		}
	}
	if v := os.Getenv(key); v != "" {
		cmd.Env = append(cmd.Env, prefix+v)
	} else {
		cmd.Env = append(cmd.Env, prefix+fallback)
	}
}

func forwardEnv(cmd *exec.Cmd, key string) {
	prefix := key + "="
	for _, e := range cmd.Env {
		if len(e) > len(prefix) && e[:len(prefix)] == prefix {
			return
		}
	}
	if v := os.Getenv(key); v != "" {
		cmd.Env = append(cmd.Env, prefix+v)
	}
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
	case <-time.After(gracefulCloseTimeout):
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
