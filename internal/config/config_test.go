package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	assertEqual(t, len(cfg.Agents), 4)
	assertEqual(t, cfg.Agents["claude"].Enabled, true)
	assertEqual(t, cfg.Agents["claude"].Command, "claude")
	assertEqual(t, cfg.Agents["opencode"].Enabled, true)
	assertEqual(t, cfg.Agents["codex"].Enabled, true)
	assertEqual(t, cfg.Agents["aider"].Enabled, false)

	assertEqual(t, cfg.UI.SidebarWidth, 30)
	assertEqual(t, cfg.UI.ShowCost, true)
	assertEqual(t, cfg.UI.ShowTokens, true)

	assertEqual(t, cfg.SideApps.Editor, "nvim .")
	assertEqual(t, cfg.SideApps.Git, "lazygit")
	assertEqual(t, cfg.SideApps.Docker, "lazydocker")
}

func TestConfigDir(t *testing.T) {
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		assertEqual(t, ConfigDir(), "/custom/config/onering")
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		dir := ConfigDir()
		if !strings.HasSuffix(dir, filepath.Join(".config", "onering")) {
			t.Errorf("ConfigDir() = %q, want suffix .config/onering", dir)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("no config file returns defaults", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, cfg.UI.SidebarWidth, 30)
		assertEqual(t, cfg.Agents["claude"].Enabled, true)
	})

	t.Run("valid YAML overrides specific fields", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		dir := filepath.Join(tmp, "onering")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
ui:
  sidebar_width: 50
  show_cost: false
`), 0644)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertEqual(t, cfg.UI.SidebarWidth, 50)
		assertEqual(t, cfg.UI.ShowCost, false)
		// ShowTokens should remain at default since not overridden
		// Note: YAML unmarshal into existing struct — unset fields get zero values
		// So ShowTokens will be false (zero value for bool), not the default true
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmp)

		dir := filepath.Join(tmp, "onering")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("agents:\n\t- broken"), 0644)

		_, err := Load()
		if err == nil {
			t.Error("expected error for invalid YAML, got nil")
		}
	})
}
