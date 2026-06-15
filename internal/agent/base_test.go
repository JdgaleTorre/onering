package agent

import "testing"

func TestBaseAgent_Name(t *testing.T) {
	a := &BaseAgent{name: "claude", command: "claude"}
	if got := a.Name(); got != "claude" {
		t.Errorf("got %q, want %q", got, "claude")
	}
}

func TestBaseAgent_Available(t *testing.T) {
	t.Run("binary exists", func(t *testing.T) {
		a := &BaseAgent{name: "go-agent", command: "go"}
		ok, _ := a.Available()
		if !ok {
			t.Error("expected go to be available")
		}
	})

	t.Run("binary does not exist", func(t *testing.T) {
		a := &BaseAgent{name: "fake", command: "nonexistent_binary_xyz_12345"}
		ok, err := a.Available()
		if ok {
			t.Error("expected nonexistent binary to be unavailable")
		}
		if err == nil {
			t.Error("expected error for nonexistent binary")
		}
	})
}
