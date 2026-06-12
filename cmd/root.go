package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/josegale/lazycode/internal/app"
	"github.com/josegale/lazycode/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lazycode",
	Short: "A lazygit-style TUI for managing code agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		m := app.New(cfg)
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
