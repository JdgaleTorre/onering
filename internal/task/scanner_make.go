package task

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

var makeTargetRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

type MakeScanner struct{}

func NewMakeScanner() *MakeScanner {
	return &MakeScanner{}
}

func (s *MakeScanner) Scan(dir string) []Task {
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
