package task

import (
	"os"
	"os/exec"
	"path/filepath"
)

type DockerScanner struct{}

func NewDockerScanner() *DockerScanner {
	return &DockerScanner{}
}

func (s *DockerScanner) Scan(dir string) []Task {
	hasDockerfile := false
	for _, name := range []string{"Dockerfile", "dockerfile", "Dockerfile.dev"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			hasDockerfile = true
			break
		}
	}
	hasCompose := false
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			hasCompose = true
			break
		}
	}
	if !hasDockerfile && !hasCompose {
		return nil
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return nil
	}

	var tasks []Task
	if hasDockerfile {
		tasks = append(tasks,
			Task{Name: "build", Command: "docker build .", Source: SourceDocker},
		)
	}
	if hasCompose {
		tasks = append(tasks,
			Task{Name: "up", Command: "docker compose up", Source: SourceDocker},
			Task{Name: "down", Command: "docker compose down", Source: SourceDocker},
			Task{Name: "logs", Command: "docker compose logs -f", Source: SourceDocker},
		)
	}
	return tasks
}
