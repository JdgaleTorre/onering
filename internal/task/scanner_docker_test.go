package task

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func dockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func TestDockerScanner_Scan(t *testing.T) {
	s := NewDockerScanner()

	if dockerAvailable() {
		t.Run("Dockerfile only", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 1 {
				t.Fatalf("got %d tasks, want 1", len(tasks))
			}
			if tasks[0].Name != "build" {
				t.Errorf("name = %q, want build", tasks[0].Name)
			}
			if tasks[0].Command != "docker build ." {
				t.Errorf("command = %q, want %q", tasks[0].Command, "docker build .")
			}
		})

		t.Run("compose.yml only", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "compose.yml"), []byte("version: '3'\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 3 {
				t.Fatalf("got %d tasks, want 3", len(tasks))
			}
			names := map[string]bool{}
			for _, task := range tasks {
				names[task.Name] = true
			}
			for _, want := range []string{"up", "down", "logs"} {
				if !names[want] {
					t.Errorf("missing task %q", want)
				}
			}
		})

		t.Run("both Dockerfile and compose", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0644)
			os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("version: '3'\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 4 {
				t.Fatalf("got %d tasks, want 4", len(tasks))
			}
		})

		t.Run("docker-compose.yaml variant", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte("version: '3'\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 3 {
				t.Fatalf("got %d tasks, want 3", len(tasks))
			}
		})

		t.Run("compose.yaml variant", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("version: '3'\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 3 {
				t.Fatalf("got %d tasks, want 3", len(tasks))
			}
		})

		t.Run("lowercase dockerfile", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "dockerfile"), []byte("FROM alpine\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 1 {
				t.Fatalf("got %d tasks, want 1", len(tasks))
			}
		})

		t.Run("Dockerfile.dev variant", func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "Dockerfile.dev"), []byte("FROM alpine\n"), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != 1 {
				t.Fatalf("got %d tasks, want 1", len(tasks))
			}
		})
	} else {
		t.Log("docker not on PATH, skipping docker-present tests")
	}

	t.Run("no docker files", func(t *testing.T) {
		dir := t.TempDir()
		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("docker not on PATH", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0644)

		t.Setenv("PATH", t.TempDir())
		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil when docker not on PATH, got %v", tasks)
		}
	})
}
