package agent

import (
	"context"
	"fmt"
	"sort"
	"testing"
)

type mockAgent struct {
	name      string
	available bool
}

func (m *mockAgent) Name() string              { return m.name }
func (m *mockAgent) Available() (bool, error) {
	if m.available {
		return true, nil
	}
	return false, fmt.Errorf("not found")
}
func (m *mockAgent) StartSession(context.Context, SessionOpts) (Session, error) { return nil, nil }
func (m *mockAgent) ResumeSession(context.Context, string) (Session, error)     { return nil, nil }

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.All()) != 0 {
		t.Errorf("expected empty registry, got %d agents", len(r.All()))
	}
}

func TestRegistry_Register_and_Get(t *testing.T) {
	r := NewRegistry()
	a := &mockAgent{name: "test", available: true}
	r.Register(a)

	t.Run("get registered agent", func(t *testing.T) {
		got, err := r.Get("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name() != "test" {
			t.Errorf("got name %q, want %q", got.Name(), "test")
		}
	})

	t.Run("get unregistered returns error", func(t *testing.T) {
		_, err := r.Get("nonexistent")
		if err == nil {
			t.Error("expected error for unregistered agent")
		}
	})

	t.Run("register overwrites", func(t *testing.T) {
		a2 := &mockAgent{name: "test", available: false}
		r.Register(a2)
		got, _ := r.Get("test")
		ok, _ := got.Available()
		if ok {
			t.Error("expected overwritten agent to be unavailable")
		}
	})
}

func TestRegistry_All(t *testing.T) {
	t.Run("returns all registered", func(t *testing.T) {
		r := NewRegistry()
		r.Register(&mockAgent{name: "a"})
		r.Register(&mockAgent{name: "b"})
		r.Register(&mockAgent{name: "c"})

		all := r.All()
		if len(all) != 3 {
			t.Fatalf("got %d agents, want 3", len(all))
		}

		names := make([]string, len(all))
		for i, a := range all {
			names[i] = a.Name()
		}
		sort.Strings(names)
		expected := []string{"a", "b", "c"}
		for i, name := range names {
			if name != expected[i] {
				t.Errorf("name[%d] = %q, want %q", i, name, expected[i])
			}
		}
	})

	t.Run("empty registry", func(t *testing.T) {
		r := NewRegistry()
		all := r.All()
		if len(all) != 0 {
			t.Errorf("expected 0 agents, got %d", len(all))
		}
	})
}

func TestRegistry_Available(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAgent{name: "yes1", available: true})
	r.Register(&mockAgent{name: "yes2", available: true})
	r.Register(&mockAgent{name: "no1", available: false})

	avail := r.Available()
	if len(avail) != 2 {
		t.Fatalf("got %d available, want 2", len(avail))
	}

	names := make(map[string]bool)
	for _, a := range avail {
		names[a.Name()] = true
	}
	if !names["yes1"] || !names["yes2"] {
		t.Errorf("expected yes1 and yes2, got %v", names)
	}
	if names["no1"] {
		t.Error("no1 should not be available")
	}
}
