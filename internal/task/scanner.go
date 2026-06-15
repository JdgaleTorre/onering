package task

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type StoredTask struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Source  string `yaml:"source"`
	Dir     string `yaml:"dir,omitempty"`
}

type TaskSource string

const (
	SourceNPM    TaskSource = "npm"
	SourcePNPM   TaskSource = "pnpm"
	SourceYarn   TaskSource = "yarn"
	SourceBun    TaskSource = "bun"
	SourceMake   TaskSource = "make"
	SourceGo     TaskSource = "go"
	SourceDocker TaskSource = "docker"
)

type Task struct {
	Name    string
	Command string
	Source  TaskSource
	Dir     string
}

func (t Task) Key() string {
	if t.Dir != "" {
		return string(t.Source) + ":" + t.Dir + "/" + t.Name
	}
	return string(t.Source) + ":" + t.Name
}

func (t Task) ToStored() StoredTask {
	return StoredTask{
		Name:    t.Name,
		Command: t.Command,
		Source:  string(t.Source),
		Dir:     t.Dir,
	}
}

func FromStored(s StoredTask) Task {
	return Task{
		Name:    s.Name,
		Command: s.Command,
		Source:  TaskSource(s.Source),
		Dir:     s.Dir,
	}
}

func TasksToStored(tasks []Task) []StoredTask {
	out := make([]StoredTask, len(tasks))
	for i, t := range tasks {
		out[i] = t.ToStored()
	}
	return out
}

func TasksFromStored(stored []StoredTask) []Task {
	out := make([]Task, len(stored))
	for i, s := range stored {
		out[i] = FromStored(s)
	}
	return out
}

type Scanner interface {
	Scan(dir string) []Task
}

func DefaultScanners(pmOverride string) []Scanner {
	return []Scanner{
		NewJSScanner(pmOverride),
		NewMakeScanner(),
		NewGoScanner(),
		NewDockerScanner(),
	}
}

func ScanTasks(dir string, pmOverride string) []Task {
	return ScanTasksWith(dir, DefaultScanners(pmOverride))
}

func ScanTasksWith(dir string, scanners []Scanner) []Task {
	var tasks []Task
	for _, s := range scanners {
		tasks = append(tasks, s.Scan(dir)...)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Source != tasks[j].Source {
			return tasks[i].Source < tasks[j].Source
		}
		if tasks[i].Dir != tasks[j].Dir {
			return tasks[i].Dir < tasks[j].Dir
		}
		return tasks[i].Name < tasks[j].Name
	})
	return tasks
}

var skipDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
}

func ScanTasksRecursive(dir string, pmOverride string) []Task {
	scanners := DefaultScanners(pmOverride)
	var tasks []Task

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && path != dir && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}

		rel := ""
		if path != dir {
			rel = strings.TrimPrefix(path, dir+string(os.PathSeparator))
		}
		for _, s := range scanners {
			for _, t := range s.Scan(path) {
				t.Dir = rel
				tasks = append(tasks, t)
			}
		}
		return nil
	})

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Source != tasks[j].Source {
			return tasks[i].Source < tasks[j].Source
		}
		if tasks[i].Dir != tasks[j].Dir {
			return tasks[i].Dir < tasks[j].Dir
		}
		return tasks[i].Name < tasks[j].Name
	})
	return tasks
}

func SortWithPreferred(tasks []Task, preferred []string) []Task {
	if len(preferred) == 0 {
		return tasks
	}
	prefIdx := make(map[string]int, len(preferred))
	for i, key := range preferred {
		prefIdx[key] = i
	}
	result := make([]Task, len(tasks))
	copy(result, tasks)
	sort.SliceStable(result, func(i, j int) bool {
		iPref, iOk := prefIdx[result[i].Key()]
		jPref, jOk := prefIdx[result[j].Key()]
		if iOk && jOk {
			return iPref < jPref
		}
		if iOk {
			return true
		}
		if jOk {
			return false
		}
		return false
	})
	return result
}
