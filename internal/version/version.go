// Package version reports the doctree binary's installed version, as
// recorded by install.sh in ~/.doctree/VERSION.
package version

import (
	"os"
	"strings"

	"github.com/foreverfl/doctree/internal/paths"
)

// Installed returns the version recorded by install.sh in ~/.doctree/VERSION,
// or "" if not recorded or unreadable.
func Installed() string {
	path, err := paths.VersionPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
