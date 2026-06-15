package task

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestJSScanner_Scan(t *testing.T) {
	t.Run("basic npm no lockfile", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"dev":"vite","build":"tsc"}}`), 0644)

		tasks := NewJSScanner("").Scan(dir)
		if len(tasks) != 3 {
			t.Fatalf("got %d tasks, want 3", len(tasks))
		}
		// install is always first
		if tasks[0].Name != "install" {
			t.Errorf("first task = %q, want install", tasks[0].Name)
		}
		if tasks[0].Command != "npm install" {
			t.Errorf("install command = %q, want %q", tasks[0].Command, "npm install")
		}
		for _, task := range tasks {
			if task.Source != SourceNPM {
				t.Errorf("source = %q, want npm", task.Source)
			}
		}
	})

	t.Run("pnpm detected via lockfile", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"dev":"vite"}}`), 0644)
		os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0644)

		tasks := NewJSScanner("").Scan(dir)
		for _, task := range tasks {
			if task.Source != SourcePNPM {
				t.Errorf("source = %q, want pnpm", task.Source)
			}
		}
	})

	t.Run("yarn detected via lockfile", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"dev":"vite"}}`), 0644)
		os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644)

		tasks := NewJSScanner("").Scan(dir)
		for _, task := range tasks {
			if task.Source != SourceYarn {
				t.Errorf("source = %q, want yarn", task.Source)
			}
		}
	})

	t.Run("bun detected via bun.lock", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"dev":"vite"}}`), 0644)
		os.WriteFile(filepath.Join(dir, "bun.lock"), []byte(""), 0644)

		tasks := NewJSScanner("").Scan(dir)
		for _, task := range tasks {
			if task.Source != SourceBun {
				t.Errorf("source = %q, want bun", task.Source)
			}
		}
	})

	t.Run("override trumps lockfile", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"dev":"vite"}}`), 0644)
		os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644)

		tasks := NewJSScanner("pnpm").Scan(dir)
		for _, task := range tasks {
			if task.Source != SourcePNPM {
				t.Errorf("source = %q, want pnpm (override)", task.Source)
			}
		}
	})

	t.Run("no package.json", func(t *testing.T) {
		dir := t.TempDir()
		tasks := NewJSScanner("").Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("empty scripts", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{}}`), 0644)

		tasks := NewJSScanner("").Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{invalid}`), 0644)

		tasks := NewJSScanner("").Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("no scripts key", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"name":"foo"}`), 0644)

		tasks := NewJSScanner("").Scan(dir)
		if tasks != nil {
			t.Errorf("expected nil, got %v", tasks)
		}
	})

	t.Run("script commands use correct pm prefix", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"),
			[]byte(`{"scripts":{"lint":"eslint ."}}`), 0644)
		os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte(""), 0644)

		tasks := NewJSScanner("").Scan(dir)
		names := map[string]string{}
		for _, task := range tasks {
			names[task.Name] = task.Command
		}
		if names["install"] != "bun install" {
			t.Errorf("install command = %q, want %q", names["install"], "bun install")
		}
		if names["lint"] != "bun run lint" {
			t.Errorf("lint command = %q, want %q", names["lint"], "bun run lint")
		}
	})
}

func TestDetectPackageManager(t *testing.T) {
	overrides := []struct {
		override string
		want     TaskSource
	}{
		{"npm", SourceNPM},
		{"pnpm", SourcePNPM},
		{"yarn", SourceYarn},
		{"bun", SourceBun},
	}
	for _, tt := range overrides {
		t.Run("override "+tt.override, func(t *testing.T) {
			got := detectPackageManager(t.TempDir(), tt.override)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	lockfiles := []struct {
		name string
		file string
		want TaskSource
	}{
		{"bun.lock", "bun.lock", SourceBun},
		{"bun.lockb", "bun.lockb", SourceBun},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", SourcePNPM},
		{"yarn.lock", "yarn.lock", SourceYarn},
	}
	for _, tt := range lockfiles {
		t.Run("lockfile "+tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, tt.file), []byte(""), 0644)
			got := detectPackageManager(dir, "")
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("no lockfiles defaults to npm", func(t *testing.T) {
		got := detectPackageManager(t.TempDir(), "")
		if got != SourceNPM {
			t.Errorf("got %q, want npm", got)
		}
	})

	t.Run("bun has priority over pnpm and yarn", func(t *testing.T) {
		dir := t.TempDir()
		for _, f := range []string{"bun.lock", "pnpm-lock.yaml", "yarn.lock"} {
			os.WriteFile(filepath.Join(dir, f), []byte(""), 0644)
		}
		got := detectPackageManager(dir, "")
		if got != SourceBun {
			t.Errorf("got %q, want bun (highest priority)", got)
		}
	})

	t.Run("unknown override falls through to lockfile detection", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte(""), 0644)
		got := detectPackageManager(dir, "unknown")
		if got != SourceYarn {
			t.Errorf("got %q, want yarn (unknown override should fall through)", got)
		}
	})

	t.Run("unknown override no lockfiles defaults to npm", func(t *testing.T) {
		got := detectPackageManager(t.TempDir(), "unknown")
		if got != SourceNPM {
			t.Errorf("got %q, want npm", got)
		}
	})
}

func TestScanTasksRecursive(t *testing.T) {
	dir := t.TempDir()

	// Root with Makefile
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build\n"), 0644)

	// Subdir with package.json
	frontend := filepath.Join(dir, "frontend")
	os.MkdirAll(frontend, 0755)
	os.WriteFile(filepath.Join(frontend, "package.json"),
		[]byte(`{"scripts":{"dev":"vite"}}`), 0644)

	// node_modules should be skipped
	nm := filepath.Join(frontend, "node_modules", "somepkg")
	os.MkdirAll(nm, 0755)
	os.WriteFile(filepath.Join(nm, "package.json"),
		[]byte(`{"scripts":{"hidden":"nope"}}`), 0644)

	// .git should be skipped
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "Makefile"), []byte("secret:\n\techo nope\n"), 0644)

	tasks := ScanTasksRecursive(dir, "")

	// Collect task keys for assertion
	keys := make(map[string]bool)
	for _, task := range tasks {
		keys[task.Key()] = true
	}

	// Root Makefile target should have empty Dir
	if !keys["make:build"] {
		t.Error("missing root make:build task")
	}

	// Frontend tasks should have Dir="frontend"
	foundFrontend := false
	for _, task := range tasks {
		if task.Dir == "frontend" && task.Source == SourceNPM {
			foundFrontend = true
			break
		}
	}
	if !foundFrontend {
		t.Error("missing frontend npm tasks")
	}

	// node_modules tasks should NOT exist
	for _, task := range tasks {
		if task.Dir == "frontend/node_modules/somepkg" {
			t.Error("should not scan node_modules")
		}
	}

	// .git tasks should NOT exist
	for _, task := range tasks {
		if task.Dir == ".git" {
			t.Error("should not scan .git")
		}
	}

	// Verify sorted
	sorted := sort.SliceIsSorted(tasks, func(i, j int) bool {
		if tasks[i].Source != tasks[j].Source {
			return tasks[i].Source < tasks[j].Source
		}
		if tasks[i].Dir != tasks[j].Dir {
			return tasks[i].Dir < tasks[j].Dir
		}
		return tasks[i].Name < tasks[j].Name
	})
	if !sorted {
		t.Error("tasks should be sorted by source, dir, name")
	}
}
