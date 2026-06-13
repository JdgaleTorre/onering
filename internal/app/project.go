package app

import (
	"os"
	"path/filepath"
	"strings"
)

// readProjectInfo returns the working directory name and the current git
// branch ("" when not in a git repo). It reads .git/HEAD directly instead
// of shelling out, so it is cheap enough to call on every refresh.
func readProjectInfo() (name, branch string) {
	wd, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	name = filepath.Base(wd)

	for dir := wd; ; dir = filepath.Dir(dir) {
		head, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
		if err == nil {
			ref := strings.TrimSpace(string(head))
			if after, ok := strings.CutPrefix(ref, "ref: refs/heads/"); ok {
				branch = after
			} else if len(ref) >= 7 {
				branch = ref[:7] // detached HEAD
			}
			return name, branch
		}
		if dir == filepath.Dir(dir) {
			return name, ""
		}
	}
}
