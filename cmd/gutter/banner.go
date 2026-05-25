//go:build !js || !wasm

package main

import (
	_ "embed"
	"fmt"
	"strings"
)

// asciiLogo is the Gutter mark drawn in ASCII. A copy lives here (alongside the
// repo-root gutter-ascii-icon.txt) because go:embed can only reach files in
// this package's directory — same arrangement as gutter.ico.
//
//go:embed gutter-ascii-icon.txt
var asciiLogo string

// printBanner prints the logo + name/version in the brand purple. Shown when
// the CLI is run with no subcommand, so a bare `gutter` greets you nicely; the
// command list follows from cobra's help.
func printBanner() {
	logo := strings.Trim(asciiLogo, "\n")
	fmt.Println(styleAccent.Render(logo))
	fmt.Println("  " + styleTitle.Render("Gutter") + styleDim.Render("  v"+version))
	fmt.Println()
}
