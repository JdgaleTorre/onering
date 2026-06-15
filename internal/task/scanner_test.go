package task

import (
	"testing"
)

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

type staticScanner struct{ tasks []Task }

func (s staticScanner) Scan(string) []Task { return s.tasks }

func TestTask_Key(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want string
	}{
		{"source+name no dir", Task{Name: "build", Source: SourceMake}, "make:build"},
		{"source+name+dir", Task{Name: "test", Source: SourceGo, Dir: "sub/pkg"}, "go:sub/pkg/test"},
		{"npm source", Task{Name: "dev", Source: SourceNPM}, "npm:dev"},
		{"pnpm source", Task{Name: "start", Source: SourcePNPM}, "pnpm:start"},
		{"bun source", Task{Name: "lint", Source: SourceBun}, "bun:lint"},
		{"empty dir treated as no dir", Task{Name: "x", Source: SourceBun, Dir: ""}, "bun:x"},
		{"docker source with dir", Task{Name: "up", Source: SourceDocker, Dir: "infra"}, "docker:infra/up"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.task.Key(), tt.want)
		})
	}
}

func TestToStored_FromStored_Roundtrip(t *testing.T) {
	tests := []Task{
		{Name: "build", Command: "make build", Source: SourceMake, Dir: ""},
		{Name: "test", Command: "go test ./...", Source: SourceGo, Dir: "sub/pkg"},
		{Name: "dev", Command: "npm run dev", Source: SourceNPM, Dir: "frontend"},
	}
	for _, orig := range tests {
		t.Run(orig.Name, func(t *testing.T) {
			stored := orig.ToStored()
			assertEqual(t, stored.Name, orig.Name)
			assertEqual(t, stored.Command, orig.Command)
			assertEqual(t, stored.Source, string(orig.Source))
			assertEqual(t, stored.Dir, orig.Dir)

			back := FromStored(stored)
			assertEqual(t, back.Name, orig.Name)
			assertEqual(t, back.Command, orig.Command)
			assertEqual(t, back.Source, orig.Source)
			assertEqual(t, back.Dir, orig.Dir)
		})
	}
}

func TestTasksToStored_TasksFromStored(t *testing.T) {
	t.Run("multiple tasks", func(t *testing.T) {
		tasks := []Task{
			{Name: "a", Command: "cmd-a", Source: SourceGo},
			{Name: "b", Command: "cmd-b", Source: SourceMake, Dir: "sub"},
		}
		stored := TasksToStored(tasks)
		assertEqual(t, len(stored), 2)

		back := TasksFromStored(stored)
		assertEqual(t, len(back), 2)
		assertEqual(t, back[0].Name, "a")
		assertEqual(t, back[1].Dir, "sub")
	})

	t.Run("empty slice", func(t *testing.T) {
		stored := TasksToStored([]Task{})
		assertEqual(t, len(stored), 0)
		back := TasksFromStored(stored)
		assertEqual(t, len(back), 0)
	})
}

func TestScanTasksWith(t *testing.T) {
	t.Run("aggregates from multiple scanners", func(t *testing.T) {
		s1 := staticScanner{tasks: []Task{
			{Name: "build", Command: "make build", Source: SourceMake},
		}}
		s2 := staticScanner{tasks: []Task{
			{Name: "test", Command: "go test", Source: SourceGo},
		}}
		result := ScanTasksWith("/tmp", []Scanner{s1, s2})
		assertEqual(t, len(result), 2)
	})

	t.Run("sorted by source then dir then name", func(t *testing.T) {
		s := staticScanner{tasks: []Task{
			{Name: "z", Source: SourceNPM},
			{Name: "a", Source: SourceNPM},
			{Name: "m", Source: SourceGo},
			{Name: "b", Source: SourceGo, Dir: "sub"},
			{Name: "a", Source: SourceGo, Dir: "sub"},
		}}
		result := ScanTasksWith("/tmp", []Scanner{s})
		assertEqual(t, result[0].Source, SourceGo)
		assertEqual(t, result[0].Name, "m")
		assertEqual(t, result[1].Source, SourceGo)
		assertEqual(t, result[1].Dir, "sub")
		assertEqual(t, result[1].Name, "a")
		assertEqual(t, result[2].Name, "b")
		assertEqual(t, result[3].Source, SourceNPM)
		assertEqual(t, result[3].Name, "a")
		assertEqual(t, result[4].Name, "z")
	})

	t.Run("empty scanner returns nil", func(t *testing.T) {
		s := staticScanner{tasks: nil}
		result := ScanTasksWith("/tmp", []Scanner{s})
		assertEqual(t, len(result), 0)
	})

	t.Run("no scanners", func(t *testing.T) {
		result := ScanTasksWith("/tmp", []Scanner{})
		assertEqual(t, len(result), 0)
	})
}

func TestSortWithPreferred(t *testing.T) {
	tasks := []Task{
		{Name: "a", Source: SourceGo},
		{Name: "b", Source: SourceMake},
		{Name: "c", Source: SourceNPM},
		{Name: "d", Source: SourceDocker},
	}

	t.Run("empty preferred unchanged", func(t *testing.T) {
		result := SortWithPreferred(tasks, nil)
		assertEqual(t, len(result), 4)
		assertEqual(t, result[0].Name, "a")
		assertEqual(t, result[3].Name, "d")
	})

	t.Run("single preferred moves to front", func(t *testing.T) {
		result := SortWithPreferred(tasks, []string{"npm:c"})
		assertEqual(t, result[0].Name, "c")
	})

	t.Run("multiple preferred ordered by preference index", func(t *testing.T) {
		result := SortWithPreferred(tasks, []string{"docker:d", "make:b"})
		assertEqual(t, result[0].Name, "d")
		assertEqual(t, result[1].Name, "b")
	})

	t.Run("preferred key not matching any task has no effect", func(t *testing.T) {
		result := SortWithPreferred(tasks, []string{"go:nonexistent"})
		assertEqual(t, result[0].Name, "a")
	})

	t.Run("non-preferred maintain relative order", func(t *testing.T) {
		result := SortWithPreferred(tasks, []string{"npm:c"})
		assertEqual(t, result[0].Name, "c")
		assertEqual(t, result[1].Name, "a")
		assertEqual(t, result[2].Name, "b")
		assertEqual(t, result[3].Name, "d")
	})

	t.Run("does not mutate original", func(t *testing.T) {
		_ = SortWithPreferred(tasks, []string{"npm:c"})
		assertEqual(t, tasks[0].Name, "a")
	})
}
