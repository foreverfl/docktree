// Package logo embeds the doctree character art used by the welcome banner.
package logo

import _ "embed"

//go:embed art.txt
var Art string