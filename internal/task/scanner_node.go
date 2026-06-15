package task

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type JSScanner struct {
	pmOverride string
}

func NewJSScanner(pmOverride string) *JSScanner {
	return &JSScanner{pmOverride: pmOverride}
}

func (s *JSScanner) Scan(dir string) []Task {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil || len(pkg.Scripts) == 0 {
		return nil
	}

	pm := detectPackageManager(dir, s.pmOverride)
	tasks := make([]Task, 0, len(pkg.Scripts)+1)
	tasks = append(tasks, Task{
		Name:    "install",
		Command: string(pm) + " install",
		Source:  pm,
	})
	for name := range pkg.Scripts {
		tasks = append(tasks, Task{
			Name:    name,
			Command: string(pm) + " run " + name,
			Source:  pm,
		})
	}
	return tasks
}

func detectPackageManager(dir string, override string) TaskSource {
	switch override {
	case "npm":
		return SourceNPM
	case "pnpm":
		return SourcePNPM
	case "yarn":
		return SourceYarn
	case "bun":
		return SourceBun
	}
	checks := []struct {
		file   string
		source TaskSource
	}{
		{"bun.lock", SourceBun},
		{"bun.lockb", SourceBun},
		{"pnpm-lock.yaml", SourcePNPM},
		{"yarn.lock", SourceYarn},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			return c.source
		}
	}
	return SourceNPM
}
