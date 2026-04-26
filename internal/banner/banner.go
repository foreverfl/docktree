// Package banner renders the doctree welcome banner shown on `doctree on`.
package banner

import (
	"fmt"
	"io"
	"strings"

	"github.com/foreverfl/doctree/assets/logo"
)

const (
	skyBlue = "\033[38;5;117m"
	reset   = "\033[0m"
)

// Print writes the doctree welcome banner to out: art on top, version label
// centered below, all wrapped in a sky-blue border. version may be empty.
func Print(out io.Writer, version string) {
	art := artLines()
	label := "doctree"
	if version != "" {
		label = "doctree " + version
	}
	width := maxWidth(art...)
	if count := runeCount(label); count > width {
		width = count
	}

	rows := []string{""}
	rows = append(rows, art...)
	rows = append(rows, "", centered(label, width), "")
	drawBox(out, rows, width)
}

// PrintLogo writes just the boxed art (no version line) to out.
func PrintLogo(out io.Writer) {
	art := artLines()
	width := maxWidth(art...)

	rows := []string{""}
	rows = append(rows, art...)
	rows = append(rows, "")
	drawBox(out, rows, width)
}

func artLines() []string {
	return strings.Split(strings.TrimRight(logo.Art, "\n"), "\n")
}

func drawBox(out io.Writer, rows []string, width int) {
	dashes := strings.Repeat("─", width+2)
	fmt.Fprintln(out, skyBlue+"╭"+dashes+"╮"+reset)
	for _, line := range rows {
		fmt.Fprintln(out, row(line, width))
	}
	fmt.Fprintln(out, skyBlue+"╰"+dashes+"╯"+reset)
}

func maxWidth(lines ...string) int {
	width := 0
	for _, line := range lines {
		if count := runeCount(line); count > width {
			width = count
		}
	}
	return width
}

// row renders one inner row: sky-blue side borders, single-space inner
// padding, content padded with spaces to width visual cells. Assumes every
// rune in content is single-width.
func row(content string, width int) string {
	pad := width - runeCount(content)
	if pad < 0 {
		pad = 0
	}
	return skyBlue + "│" + reset +
		" " + content + strings.Repeat(" ", pad) + " " +
		skyBlue + "│" + reset
}

func centered(text string, width int) string {
	count := runeCount(text)
	if count >= width {
		return text
	}
	return strings.Repeat(" ", (width-count)/2) + text
}

func runeCount(text string) int {
	count := 0
	for range text {
		count++
	}
	return count
}