package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoScanner_Scan(t *testing.T) {
	s := NewGoScanner()

	t.Run("has go.mod", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)

		tasks := s.Scan(dir)
		if len(tasks) != 4 {
			t.Fatalf("got %d tasks, want 4", len(tasks))
		}

		expected := []struct {
			name    string
			command string
		}{
			{"build", "go build ./..."},
			{"test", "go test ./..."},
			{"vet", "go vet ./..."},
			{"fmt", "go fmt ./..."},
		}
		for i, e := range expected {
			if tasks[i].Name != e.name {
				t.Errorf("task[%d].Name = %q, want %q", i, tasks[i].Name, e.name)
			}
			if tasks[i].Command != e.command {
				t.Errorf("task[%d].Command = %q, want %q", i, tasks[i].Command, e.command)
			}
			if tasks[i].Source != SourceGo {
				t.Errorf("task[%d].Source = %q, want go", i, tasks[i].Source)
			}
		}
	})

	t.Run("no go.mod", func(t *testing.T) {
		dir := t.TempDir()
		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})
}
