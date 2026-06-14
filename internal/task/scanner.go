package task

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

type TaskSource string

const (
	SourceNPM  TaskSource = "npm"
	SourcePNPM TaskSource = "pnpm"
	SourceYarn TaskSource = "yarn"
	SourceBun  TaskSource = "bun"
	SourceMake TaskSource = "make"
	SourceGo   TaskSource = "go"
)

type Task struct {
	Name    string
	Command string
	Source  TaskSource
}

func (t Task) Key() string {
	return string(t.Source) + ":" + t.Name
}

func ScanTasks(dir string, pmOverride string) []Task {
	var tasks []Task
	tasks = append(tasks, parsePackageJSON(dir, pmOverride)...)
	tasks = append(tasks, parseMakefile(dir)...)
	tasks = append(tasks, parseGoMod(dir)...)
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Source != tasks[j].Source {
			return tasks[i].Source < tasks[j].Source
		}
		return tasks[i].Name < tasks[j].Name
	})
	return tasks
}

func parsePackageJSON(dir string, pmOverride string) []Task {
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

	pm := detectPackageManager(dir, pmOverride)
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

var makeTargetRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

func parseMakefile(dir string) []Task {
	f, err := os.Open(filepath.Join(dir, "Makefile"))
	if err != nil {
		return nil
	}
	defer f.Close()

	var tasks []Task
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		m := makeTargetRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		target := m[1]
		tasks = append(tasks, Task{
			Name:    target,
			Command: "make " + target,
			Source:  SourceMake,
		})
	}
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

func parseGoMod(dir string) []Task {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return nil
	}
	if _, err := exec.LookPath("go"); err != nil {
		return nil
	}
	return []Task{
		{Name: "build", Command: "go build ./...", Source: SourceGo},
		{Name: "test", Command: "go test ./...", Source: SourceGo},
		{Name: "vet", Command: "go vet ./...", Source: SourceGo},
		{Name: "fmt", Command: "go fmt ./...", Source: SourceGo},
	}
}
