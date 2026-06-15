package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/josegale/onering/internal/task"
)

func TestState_IsPreferred(t *testing.T) {
	t.Run("returns true when present", func(t *testing.T) {
		s := &State{PreferredTasks: map[string][]string{"/proj": {"go:build"}}}
		assertEqual(t, s.IsPreferred("/proj", "go:build"), true)
	})

	t.Run("returns false when absent", func(t *testing.T) {
		s := &State{PreferredTasks: map[string][]string{"/proj": {"go:build"}}}
		assertEqual(t, s.IsPreferred("/proj", "go:test"), false)
	})

	t.Run("returns false for unknown project", func(t *testing.T) {
		s := &State{PreferredTasks: map[string][]string{"/proj": {"go:build"}}}
		assertEqual(t, s.IsPreferred("/other", "go:build"), false)
	})

	t.Run("nil map does not panic", func(t *testing.T) {
		s := &State{}
		assertEqual(t, s.IsPreferred("/proj", "go:build"), false)
	})
}

func TestState_TogglePreferred(t *testing.T) {
	t.Run("add to empty", func(t *testing.T) {
		s := &State{}
		s.TogglePreferred("/proj", "go:build")
		assertEqual(t, s.IsPreferred("/proj", "go:build"), true)
	})

	t.Run("add to existing list", func(t *testing.T) {
		s := &State{PreferredTasks: map[string][]string{"/proj": {"go:build"}}}
		s.TogglePreferred("/proj", "go:test")
		assertEqual(t, s.IsPreferred("/proj", "go:build"), true)
		assertEqual(t, s.IsPreferred("/proj", "go:test"), true)
	})

	t.Run("remove existing", func(t *testing.T) {
		s := &State{PreferredTasks: map[string][]string{"/proj": {"go:build", "go:test"}}}
		s.TogglePreferred("/proj", "go:build")
		assertEqual(t, s.IsPreferred("/proj", "go:build"), false)
		assertEqual(t, s.IsPreferred("/proj", "go:test"), true)
	})

	t.Run("toggle on then off roundtrip", func(t *testing.T) {
		s := &State{}
		s.TogglePreferred("/proj", "go:build")
		assertEqual(t, s.IsPreferred("/proj", "go:build"), true)
		s.TogglePreferred("/proj", "go:build")
		assertEqual(t, s.IsPreferred("/proj", "go:build"), false)
	})
}

func TestState_RecordProject(t *testing.T) {
	t.Run("prepends new project", func(t *testing.T) {
		s := &State{RecentProjects: []string{"/old"}}
		s.RecordProject("/new")
		assertEqual(t, s.RecentProjects[0], "/new")
		assertEqual(t, len(s.RecentProjects), 2)
	})

	t.Run("deduplicates existing project", func(t *testing.T) {
		s := &State{RecentProjects: []string{"/a", "/b", "/c"}}
		s.RecordProject("/b")
		assertEqual(t, s.RecentProjects[0], "/b")
		assertEqual(t, len(s.RecentProjects), 3)
	})

	t.Run("caps at maxRecentProjects", func(t *testing.T) {
		projects := make([]string, maxRecentProjects)
		for i := range projects {
			projects[i] = "/proj" + string(rune('A'+i%26))
		}
		s := &State{RecentProjects: projects}
		s.RecordProject("/newest")
		assertEqual(t, len(s.RecentProjects), maxRecentProjects)
		assertEqual(t, s.RecentProjects[0], "/newest")
	})

	t.Run("empty initial state", func(t *testing.T) {
		s := &State{}
		s.RecordProject("/first")
		assertEqual(t, len(s.RecentProjects), 1)
		assertEqual(t, s.RecentProjects[0], "/first")
	})
}

func TestState_RemoveProject(t *testing.T) {
	t.Run("removes existing", func(t *testing.T) {
		s := &State{RecentProjects: []string{"/a", "/b", "/c"}}
		s.RemoveProject("/b")
		assertEqual(t, len(s.RecentProjects), 2)
		for _, p := range s.RecentProjects {
			if p == "/b" {
				t.Error("project /b should have been removed")
			}
		}
	})

	t.Run("no-op for missing", func(t *testing.T) {
		s := &State{RecentProjects: []string{"/a"}}
		s.RemoveProject("/missing")
		assertEqual(t, len(s.RecentProjects), 1)
	})

	t.Run("works with empty list", func(t *testing.T) {
		s := &State{RecentProjects: []string{}}
		s.RemoveProject("/x")
		assertEqual(t, len(s.RecentProjects), 0)
	})
}

func TestState_ProjectTasks(t *testing.T) {
	t.Run("load from nil map", func(t *testing.T) {
		s := &State{}
		got := s.LoadProjectTasks("/proj")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("load unknown dir", func(t *testing.T) {
		s := &State{ProjectTasks: map[string][]task.StoredTask{}}
		got := s.LoadProjectTasks("/unknown")
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		s := &State{}
		tasks := []task.StoredTask{
			{Name: "build", Command: "go build", Source: "go"},
		}
		s.SaveProjectTasks("/proj", tasks)

		got := s.LoadProjectTasks("/proj")
		assertEqual(t, len(got), 1)
		assertEqual(t, got[0].Name, "build")
	})

	t.Run("save creates map if nil", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		s := &State{}
		s.SaveProjectTasks("/proj", []task.StoredTask{{Name: "x"}})
		if s.ProjectTasks == nil {
			t.Error("ProjectTasks should be initialized")
		}
	})

	t.Run("save overwrites existing", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		s := &State{ProjectTasks: map[string][]task.StoredTask{
			"/proj": {{Name: "old"}},
		}}
		s.SaveProjectTasks("/proj", []task.StoredTask{{Name: "new"}})
		assertEqual(t, s.ProjectTasks["/proj"][0].Name, "new")
	})
}

func TestLoadState(t *testing.T) {
	t.Run("no state file returns empty state", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		s := LoadState()
		if s == nil {
			t.Fatal("expected non-nil state")
		}
		// When no file exists, os.ReadFile fails and LoadState returns
		// a zero-initialized State (nil slices/maps). The nil-safety
		// initialization only runs after a successful Unmarshal.
		assertEqual(t, len(s.RecentProjects), 0)
	})

	t.Run("valid state YAML", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		dir := filepath.Join(tmp, "onering")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(`
recent_projects:
  - /proj/a
  - /proj/b
preferred_tasks:
  /proj/a:
    - "go:build"
`), 0644)

		s := LoadState()
		assertEqual(t, len(s.RecentProjects), 2)
		assertEqual(t, s.RecentProjects[0], "/proj/a")
		assertEqual(t, s.IsPreferred("/proj/a", "go:build"), true)
	})

	t.Run("malformed YAML returns empty state", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		dir := filepath.Join(tmp, "onering")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(":::bad:::"), 0644)

		s := LoadState()
		if s == nil {
			t.Fatal("expected non-nil state even with bad YAML")
		}
	})
}

func TestState_Save(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	s := &State{
		RecentProjects: []string{"/proj"},
		PreferredTasks: map[string][]string{"/proj": {"go:build"}},
	}
	err := s.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "onering", "state.yaml"))
	if err != nil {
		t.Fatalf("read saved state: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved state file is empty")
	}
}
