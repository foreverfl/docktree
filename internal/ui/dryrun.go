package ui

import (
	"fmt"
	"io"
)

// DryRunf prints "[dry-run] " followed by the formatted message. Commands
// invoked with --dry-run use this to preview each side effect they would
// have performed. The caller supplies any trailing newline so multi-line
// previews stay readable.
func DryRunf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[dry-run] "+format, args...)
}
