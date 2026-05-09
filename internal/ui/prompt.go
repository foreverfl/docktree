// Package ui contains raw interactive prompt primitives used across gitt's
// commands. Confirm is a small bufio-based yes/no prompt; Select wraps
// charmbracelet/huh to give arrow-key navigation, a coloured selection
// indicator, and a checkmark on the chosen entry. Both functions detect
// non-terminals up front and return ErrNoTTY so callers can translate that
// into a flag-driven hint (--yes, --project-type, …) rather than blocking on
// a closed stdin.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
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
	if file, ok := in.(*os.File); ok {
		stat, err := file.Stat()
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

// Option is one entry in a Select prompt. Disabled options are still rendered
// (so users can see what's coming) but cannot be picked — the underlying huh
// form runs Validate on each submission and returns the option's Note as the
// rejection message, leaving the cursor in place so the user can pick again.
//
// The same struct is intended to back a future MultiSelect helper: huh's
// Option[T] type is what both NewSelect and NewMultiSelect consume, so adding
// a list-returning variant is a matter of swapping the constructor.
type Option struct {
	Label    string
	Value    string
	Disabled bool
	Note     string
}

// Select asks the user to pick one of the given options using arrow keys and
// Enter, and returns the chosen option's Value. defaultIndex positions the
// cursor on initial render. Returns ErrNoTTY when stdin is not a terminal so
// callers can advise on flag-based bypass (--yes, --project-type, …).
func Select(message string, options []Option, defaultIndex int) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}
	if defaultIndex < 0 || defaultIndex >= len(options) {
		return "", fmt.Errorf("default index %d out of range", defaultIndex)
	}
	if options[defaultIndex].Disabled {
		return "", fmt.Errorf("default option %q is disabled", options[defaultIndex].Label)
	}

	if !isTerminal(os.Stdin) {
		return "", ErrNoTTY
	}

	huhOptions, disabledByValue := buildHuhOptions(options)
	chosen := options[defaultIndex].Value

	err := huh.NewSelect[string]().
		Title(message).
		Options(huhOptions...).
		Value(&chosen).
		Validate(func(v string) error {
			if opt, blocked := disabledByValue[v]; blocked {
				note := opt.Note
				if note == "" {
					note = "not available yet"
				}
				return fmt.Errorf("%q is %s — pick another", opt.Label, note)
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return chosen, nil
}

// isTerminal reports whether the given file is connected to an interactive
// terminal. Plain os.ModeCharDevice checks misclassify /dev/null on macOS
// (it is itself a character device), so we defer to go-isatty which queries
// the kernel via tcgetattr — only real ttys answer.
func isTerminal(file *os.File) bool {
	fd := file.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// buildHuhOptions converts gitt's Option slice into huh.Option[string] entries
// and returns a lookup of values that should be rejected by Validate. Disabled
// entries are decorated with an "(unavailable)" suffix and any Note text so
// the preview is visible without being selectable.
func buildHuhOptions(options []Option) ([]huh.Option[string], map[string]Option) {
	huhOptions := make([]huh.Option[string], len(options))
	disabledByValue := make(map[string]Option)
	for index, option := range options {
		label := option.Label
		if option.Note != "" {
			label += " — " + option.Note
		}
		if option.Disabled {
			label += " (unavailable)"
			disabledByValue[option.Value] = option
		}
		huhOptions[index] = huh.NewOption(label, option.Value)
	}
	return huhOptions, disabledByValue
}
