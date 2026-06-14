package app

import (
	"os"
	"path/filepath"
	"strings"
)

// readProjectInfo returns the directory name and the current git branch
// ("" when not in a git repo). It reads .git/HEAD directly instead of
// shelling out, so it is cheap enough to call on every refresh.
func readProjectInfo(dir string) (name, branch string) {
	name = filepath.Base(dir)

	for d := dir; ; d = filepath.Dir(d) {
		head, err := os.ReadFile(filepath.Join(d, ".git", "HEAD"))
		if err == nil {
			ref := strings.TrimSpace(string(head))
			if after, ok := strings.CutPrefix(ref, "ref: refs/heads/"); ok {
				branch = after
			} else if len(ref) >= 7 {
				branch = ref[:7] // detached HEAD
			}
			return name, branch
		}
		if d == filepath.Dir(d) {
			return name, ""
		}
	}
}
