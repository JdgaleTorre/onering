package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoonScanner_Scan(t *testing.T) {
	s := NewMoonScanner()

	t.Run("no moon.yml", func(t *testing.T) {
		dir := t.TempDir()
		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("moon.yml with tasks", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte(`
tasks:
  build:
    command: tsc
  test:
    command: vitest run
`), 0644)

		tasks := s.Scan(dir)
		if len(tasks) != 2 {
			t.Fatalf("got %d tasks, want 2", len(tasks))
		}

		names := map[string]bool{}
		for _, task := range tasks {
			names[task.Name] = true
			if task.Source != SourceMoon {
				t.Errorf("task %q source = %q, want %q", task.Name, task.Source, SourceMoon)
			}
		}
		for _, want := range []string{"build", "test"} {
			if !names[want] {
				t.Errorf("missing task %q", want)
			}
		}
	})

	t.Run("slug derived from directory name", func(t *testing.T) {
		projectDir := t.TempDir()
		os.WriteFile(filepath.Join(projectDir, "moon.yml"), []byte(`
tasks:
  build:
    command: tsc
`), 0644)
		slug := filepath.Base(projectDir)

		tasks := s.Scan(projectDir)
		if len(tasks) != 1 {
			t.Fatalf("got %d tasks, want 1", len(tasks))
		}
		want := "moon run " + slug + ":build"
		if tasks[0].Command != want {
			t.Errorf("command = %q, want %q", tasks[0].Command, want)
		}
	})

	t.Run("moon.yml with no tasks key", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte(`
project:
  name: my-app
`), 0644)

		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("moon.yml with empty tasks", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte("tasks: {}\n"), 0644)

		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("tasks without command field", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte(`
tasks:
  deploy:
    dependsOn:
      - build
      - test
`), 0644)

		tasks := s.Scan(dir)
		if len(tasks) != 1 {
			t.Fatalf("got %d tasks, want 1", len(tasks))
		}
		if tasks[0].Name != "deploy" {
			t.Errorf("name = %q, want deploy", tasks[0].Name)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte("{{invalid\n"), 0644)

		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("multiple tasks with correct commands", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "moon.yml"), []byte(`
tasks:
  lint:
    command: eslint .
  typecheck:
    command: tsc --noEmit
  test:
    command: vitest run
`), 0644)
		slug := filepath.Base(dir)

		tasks := s.Scan(dir)
		if len(tasks) != 3 {
			t.Fatalf("got %d tasks, want 3", len(tasks))
		}

		cmds := map[string]string{}
		for _, task := range tasks {
			cmds[task.Name] = task.Command
		}
		if cmd := cmds["lint"]; cmd != "moon run "+slug+":lint" {
			t.Errorf("lint command = %q, want %q", cmd, "moon run "+slug+":lint")
		}
		if cmd := cmds["test"]; cmd != "moon run "+slug+":test" {
			t.Errorf("test command = %q", cmd)
		}
		if cmd := cmds["typecheck"]; cmd != "moon run "+slug+":typecheck" {
			t.Errorf("typecheck command = %q", cmd)
		}
	})
}
