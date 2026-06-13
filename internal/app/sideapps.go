package app

import (
	"os/exec"
	"strings"

	"github.com/josegale/lazycode/internal/agent"
	"github.com/josegale/lazycode/internal/config"
	"github.com/josegale/lazycode/internal/terminal"
	"github.com/josegale/lazycode/internal/util"
)

// SideApp is an embeddable side application (editor, lazygit, ...) defined
// in the config. Sess is nil while the app is not running.
type SideApp struct {
	Name string
	Cmd  string
	Sess agent.Session
}

func (a SideApp) Running() bool {
	return a.Sess != nil && a.Sess.State() == agent.StateRunning
}

func buildSideApps(cfg *config.Config) []SideApp {
	var apps []SideApp
	if cfg.SideApps.Editor != "" {
		apps = append(apps, SideApp{Name: "editor", Cmd: cfg.SideApps.Editor})
	}
	if cfg.SideApps.Git != "" {
		apps = append(apps, SideApp{Name: "git", Cmd: cfg.SideApps.Git})
	}
	return apps
}

func startSideApp(app SideApp) (agent.Session, error) {
	parts := strings.Fields(app.Cmd)
	cmd := exec.Command(parts[0], parts[1:]...)

	ptyHandle, err := terminal.StartPTY(cmd)
	if err != nil {
		return nil, err
	}

	// Empty exit input: side apps are killed on Close rather than asked
	// to quit, since there is no universal quit command.
	return agent.NewPTYSession(ptyHandle, app.Name, app.Cmd, util.NewID(), ""), nil
}
