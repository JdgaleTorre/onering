package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/josegale/lazycode/internal/task"
)

// StoredTask is re-exported for YAML serialization.
type StoredTask = task.StoredTask

const maxRecentProjects = 50

type State struct {
	RecentProjects []string              `yaml:"recent_projects"`
	PreferredTasks map[string][]string   `yaml:"preferred_tasks,omitempty"`
	ProjectTasks   map[string][]StoredTask `yaml:"project_tasks,omitempty"`
}

func statePath() string {
	return filepath.Join(ConfigDir(), "state.yaml")
}

func LoadState() *State {
	s := &State{}
	data, err := os.ReadFile(statePath())
	if err != nil {
		return s
	}
	yaml.Unmarshal(data, s)
	if s.RecentProjects == nil {
		s.RecentProjects = []string{}
	}
	if s.PreferredTasks == nil {
		s.PreferredTasks = make(map[string][]string)
	}
	if s.ProjectTasks == nil {
		s.ProjectTasks = make(map[string][]StoredTask)
	}
	return s
}

func (s *State) SaveProjectTasks(dir string, tasks []StoredTask) {
	if s.ProjectTasks == nil {
		s.ProjectTasks = make(map[string][]StoredTask)
	}
	s.ProjectTasks[dir] = tasks
	s.Save()
}

func (s *State) LoadProjectTasks(dir string) []StoredTask {
	if s.ProjectTasks == nil {
		return nil
	}
	return s.ProjectTasks[dir]
}

func (s *State) Save() error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	dir := ConfigDir()
	os.MkdirAll(dir, 0755)
	return os.WriteFile(statePath(), data, 0644)
}

func (s *State) RemoveProject(dir string) {
	filtered := make([]string, 0, len(s.RecentProjects))
	for _, p := range s.RecentProjects {
		if p != dir {
			filtered = append(filtered, p)
		}
	}
	s.RecentProjects = filtered
}

func (s *State) IsPreferred(projectDir, taskKey string) bool {
	for _, k := range s.PreferredTasks[projectDir] {
		if k == taskKey {
			return true
		}
	}
	return false
}

func (s *State) TogglePreferred(projectDir, taskKey string) {
	if s.PreferredTasks == nil {
		s.PreferredTasks = make(map[string][]string)
	}
	keys := s.PreferredTasks[projectDir]
	for i, k := range keys {
		if k == taskKey {
			s.PreferredTasks[projectDir] = append(keys[:i], keys[i+1:]...)
			return
		}
	}
	s.PreferredTasks[projectDir] = append(keys, taskKey)
}

func (s *State) RecordProject(dir string) {
	filtered := make([]string, 0, len(s.RecentProjects))
	for _, p := range s.RecentProjects {
		if p != dir {
			filtered = append(filtered, p)
		}
	}
	s.RecentProjects = append([]string{dir}, filtered...)
	if len(s.RecentProjects) > maxRecentProjects {
		s.RecentProjects = s.RecentProjects[:maxRecentProjects]
	}
}
