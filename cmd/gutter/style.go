//go:build !js || !wasm

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	styleOK     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	styleWarn   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	styleErr    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	styleInfo   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Bold(true)
)

func printTitle(s string)              { fmt.Println(styleTitle.Render("▶ " + s)) }
func printOK(format string, a ...any)  { fmt.Println(styleOK.Render("✓") + " " + fmt.Sprintf(format, a...)) }
func printWarn(format string, a ...any) {
	fmt.Println(styleWarn.Render("!") + " " + fmt.Sprintf(format, a...))
}
func printInfo(format string, a ...any) {
	fmt.Println(styleInfo.Render("·") + " " + fmt.Sprintf(format, a...))
}
func printDim(format string, a ...any) { fmt.Println(styleDim.Render(fmt.Sprintf(format, a...))) }
func printErr(format string, a ...any) {
	fmt.Fprintln(os.Stderr, styleErr.Render("✗")+" "+fmt.Sprintf(format, a...))
}

func bail(format string, a ...any) {
	printErr(format, a...)
	os.Exit(1)
}
