package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JdgaleTorre/onering/internal/app"
	"github.com/JdgaleTorre/onering/internal/config"
	"github.com/JdgaleTorre/onering/internal/terminal"
	"github.com/spf13/cobra"
)

var Version = ""

func version() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

var rootCmd = &cobra.Command{
Use: "onering",
	Short: "A lazygit-style TUI for managing code agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Detect the host terminal's colors before Bubbletea takes over stdin,
		// so embedded child terminals can be made to match.
		terminal.DetectHostColors()

		m := app.New(cfg, version())
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
