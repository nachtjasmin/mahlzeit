// Package icons is a convenient wrapper for embedding the SVG icons into the binary.
package icons

import "embed"

//go:embed "*.svg"
var Icons embed.FS
