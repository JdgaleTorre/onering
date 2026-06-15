package task

import (
	"os"
	"os/exec"
	"path/filepath"
)

type GoScanner struct{}

func NewGoScanner() *GoScanner {
	return &GoScanner{}
}

func (s *GoScanner) Scan(dir string) []Task {
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
