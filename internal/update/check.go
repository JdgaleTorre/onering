package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type AvailableMsg struct {
	Version string
}

type CopiedMsg struct{}

type ClearCopiedMsg struct{}

const (
	repo       = "JdgaleTorre/onering"
	installCmd = "go install github.com/JdgaleTorre/onering@latest"
)

type release struct {
	TagName string `json:"tag_name"`
}

func Check(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		if currentVersion == "dev" || currentVersion == "" {
			return nil
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil
		}

		var r release
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil
		}

		latest := strings.TrimPrefix(r.TagName, "v")
		if isNewer(latest, currentVersion) {
			return AvailableMsg{Version: latest}
		}
		return nil
	}
}

func CopyInstallCmd() tea.Cmd {
	return func() tea.Msg {
		if err := toClipboard(installCmd); err != nil {
			return nil
		}
		return CopiedMsg{}
	}
}

func isNewer(latest, current string) bool {
	l := parseVersion(latest)
	c := parseVersion(current)
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	parts := strings.SplitN(strings.TrimPrefix(v, "v"), ".", 3)
	var result [3]int
	for i := range parts {
		if i >= 3 {
			break
		}
		numStr, _, _ := strings.Cut(parts[i], "-")
		result[i], _ = strconv.Atoi(numStr)
	}
	return result
}

func toClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found")
		}
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
