package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrNoTTY signals that stdin is not a terminal, so an interactive prompt
// would block on a closed/piped stream. Callers should translate this into
// a user-facing error like "use --yes to bypass confirmation".
var ErrNoTTY = errors.New("stdin is not a terminal")

const maxAttempts = 3

// Confirm asks a yes/no question on stdin/stderr and returns the user's
// answer. defaultYes controls both the [Y/n] vs [y/N] hint and the value
// returned when the user just hits enter.
func Confirm(message string, defaultYes bool) (bool, error) {
	return confirm(os.Stdin, os.Stderr, message, defaultYes)
}

func confirm(in io.Reader, out io.Writer, message string, defaultYes bool) (bool, error) {
	if f, ok := in.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false, err
		}
		if stat.Mode()&os.ModeCharDevice == 0 {
			return false, ErrNoTTY
		}
	}

	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	reader := bufio.NewReader(in)
	for range maxAttempts {
		if _, err := fmt.Fprintf(out, "%s %s ", message, hint); err != nil {
			return false, err
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, ErrNoTTY
			}
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(out, "please answer y or n")
	}
	return false, fmt.Errorf("too many invalid responses")
}
