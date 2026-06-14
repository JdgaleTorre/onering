package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type State struct {
	RecentProjects []string `yaml:"recent_projects"`
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
	return s
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

func (s *State) RecordProject(dir string) {
	filtered := make([]string, 0, len(s.RecentProjects))
	for _, p := range s.RecentProjects {
		if p != dir {
			filtered = append(filtered, p)
		}
	}
	s.RecentProjects = append([]string{dir}, filtered...)
	if len(s.RecentProjects) > 50 {
		s.RecentProjects = s.RecentProjects[:50]
	}
}
