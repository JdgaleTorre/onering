package app

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/josegale/onering/internal/agent"
	"github.com/josegale/onering/internal/config"
	"github.com/josegale/onering/internal/terminal"
	"github.com/josegale/onering/internal/util"
)

type SideApp struct {
	Name      string
	Cmd       string
	Installed bool
	Sess      agent.Session
}

func (a SideApp) Running() bool {
	return a.Sess != nil && a.Sess.State() == agent.StateRunning
}

type appDef struct {
	Name  string
	Cmd   string
	Known bool
}

var knownApps = map[string]appDef{
	"editor": {Name: "editor", Cmd: "nvim .", Known: true},
	"git":    {Name: "git", Cmd: "lazygit", Known: true},
	"docker": {Name: "docker", Cmd: "lazydocker", Known: true},
}

func buildSideApps(cfg *config.Config) []SideApp {
	var apps []SideApp

	addApp := func(name, cmd string) {
		enabled := true
		if v, ok := cfg.SideApps.Enable[name]; ok {
			enabled = v
		}
		if !enabled || cmd == "" {
			return
		}
		binary := strings.Fields(cmd)[0]
		_, err := exec.LookPath(binary)
		apps = append(apps, SideApp{
			Name:      name,
			Cmd:       cmd,
			Installed: err == nil,
		})
	}

	for _, def := range knownApps {
		cmd := def.Cmd
		switch def.Name {
		case "editor":
			if cfg.SideApps.Editor != "" {
				cmd = cfg.SideApps.Editor
			}
		case "git":
			if cfg.SideApps.Git != "" {
				cmd = cfg.SideApps.Git
			}
		case "docker":
			if cfg.SideApps.Docker != "" {
				cmd = cfg.SideApps.Docker
			}
		}
		addApp(def.Name, cmd)
	}

	for _, e := range cfg.SideApps.Extra {
		addApp(e.Name, e.Command)
	}

	return apps
}

func installHint(name, cmd string) string {
	urls := map[string]string{
		"nvim":       "https://neovim.io/",
		"lazygit":    "https://github.com/jesseduffield/lazygit",
		"lazydocker": "https://github.com/jesseduffield/lazydocker",
	}

	binary := strings.Fields(cmd)[0]
	url := urls[binary]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s (%s) is not installed\n\n", name, binary))
	if url != "" {
		b.WriteString(fmt.Sprintf("  Visit: %s\n", url))
	}
	b.WriteString(fmt.Sprintf("\nInstall with your package manager:\n"))
	b.WriteString(fmt.Sprintf("  brew install %s\n", binary))
	b.WriteString(fmt.Sprintf("  sudo apt install %s", binary))
	return b.String()
}

func startSideApp(app SideApp, dir string) (agent.Session, error) {
	parts := strings.Fields(app.Cmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir

	ptyHandle, err := terminal.StartPTY(cmd)
	if err != nil {
		return nil, err
	}

	return agent.NewPTYSession(ptyHandle, app.Name, app.Cmd, util.NewID(), ""), nil
}
