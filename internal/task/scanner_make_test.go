package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeScanner_Scan(t *testing.T) {
	s := NewMakeScanner()

	tests := []struct {
		name     string
		content  string
		wantLen  int
		wantName []string
	}{
		{
			name:     "basic targets",
			content:  "build:\n\tgo build\ntest:\n\tgo test\n",
			wantLen:  2,
			wantName: []string{"build", "test"},
		},
		{
			name:     "target with deps",
			content:  "all: build test\n",
			wantLen:  1,
			wantName: []string{"all"},
		},
		{
			name:     "hyphens in target",
			content:  "my-target:\n\techo hi\n",
			wantLen:  1,
			wantName: []string{"my-target"},
		},
		{
			name:     "underscores in target",
			content:  "my_target:\n\techo hi\n",
			wantLen:  1,
			wantName: []string{"my_target"},
		},
		{
			name:    "ignore variables",
			content: "FOO = bar\n",
			wantLen: 0,
		},
		{
			name:    "ignore .PHONY",
			content: ".PHONY: build\n",
			wantLen: 0,
		},
		{
			name:    "ignore indented lines",
			content: "\tbuild:\n",
			wantLen: 0,
		},
		{
			name:    "digit-starting name rejected",
			content: "1foo:\n",
			wantLen: 0,
		},
		{
			name:     "underscore-starting name accepted",
			content:  "_foo:\n",
			wantLen:  1,
			wantName: []string{"_foo"},
		},
		{
			name:    "empty Makefile",
			content: "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "Makefile"), []byte(tt.content), 0644)

			tasks := s.Scan(dir)
			if len(tasks) != tt.wantLen {
				t.Fatalf("got %d tasks, want %d", len(tasks), tt.wantLen)
			}
			for i, name := range tt.wantName {
				if tasks[i].Name != name {
					t.Errorf("task[%d].Name = %q, want %q", i, tasks[i].Name, name)
				}
				if tasks[i].Source != SourceMake {
					t.Errorf("task[%d].Source = %q, want %q", i, tasks[i].Source, SourceMake)
				}
				if tasks[i].Command != "make "+name {
					t.Errorf("task[%d].Command = %q, want %q", i, tasks[i].Command, "make "+name)
				}
			}
		})
	}

	t.Run("no Makefile", func(t *testing.T) {
		dir := t.TempDir()
		tasks := s.Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})
}
