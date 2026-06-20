package task

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MoonScanner struct{}

func NewMoonScanner() *MoonScanner {
	return &MoonScanner{}
}

type moonConfig struct {
	Tasks map[string]moonTask `yaml:"tasks"`
}

type moonTask struct {
	_ struct{} `yaml:",inline"`
}

func (s *MoonScanner) Scan(dir string) []Task {
	data, err := os.ReadFile(filepath.Join(dir, "moon.yml"))
	if err != nil {
		return nil
	}

	var cfg moonConfig
	if yaml.Unmarshal(data, &cfg) != nil || len(cfg.Tasks) == 0 {
		return nil
	}

	slug := filepath.Base(dir)
	tasks := make([]Task, 0, len(cfg.Tasks))
	for name := range cfg.Tasks {
		tasks = append(tasks, Task{
			Name:    name,
			Command: "moon run " + slug + ":" + name,
			Source:  SourceMoon,
		})
	}
	return tasks
}
