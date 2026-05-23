//go:build !js || !wasm

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const mainGoTemplate = `package main

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

type App struct{}

func (App) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title: "__NAME__",
		Theme: themes.Apple,
		AppBar: widgets.AppBar{
			Title: "__NAME__",
		},
		Body: widgets.Surface{
			Variant: widgets.SurfaceAlt,
			Child: widgets.Center{
				Child: widgets.Card{
					Variant: widgets.CardFeature,
					Child: widgets.Column{
						CrossAxisAlign: widgets.CrossAxisCenter,
						Spacing:        16,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H2, Text: "Hello, Gutter!"},
							widgets.Body{Text: "Pick a theme and ship — no CSS needed."},
							widgets.Button{
								Variant: widgets.ButtonPrimary,
								Label:   "Get started",
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	gutter.RunApp(App{})
}
`

const indexHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>__NAME__</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Lexend:wght@100..900&display=swap" rel="stylesheet">
  <style>
    html, body { margin: 0; padding: 0; width: 100%; height: 100%; font-family: Lexend, system-ui, sans-serif; }
    #app { width: 100%; height: 100%; }
  </style>
</head>
<body>
  <div id="app"></div>
  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject).then((result) => {
      go.run(result.instance);
    });
  </script>
</body>
</html>
`

const goModTemplate = `module __MODULE__

go 1.21

require github.com/Runway-Club/gutter v0.0.0
`

var (
	nameRE   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	moduleRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)
)

func newCmd() *cobra.Command {
	var modulePath string
	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new gutter project",
		Long:  "Scaffold a new gutter project. Without arguments, prompts interactively for project name and module path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			return runNew(name, modulePath)
		},
	}
	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (e.g. github.com/you/project)")
	return cmd
}

func runNew(name, modulePath string) error {
	// Interactive mode: prompt for any missing values.
	if name == "" || modulePath == "" {
		fields := []huh.Field{}
		if name == "" {
			fields = append(fields, huh.NewInput().
				Title("Project name").
				Description("Used as the directory and binary name").
				Placeholder("my-app").
				Value(&name).
				Validate(validateName))
		}
		if modulePath == "" {
			defaultMod := name
			modulePath = defaultMod
			fields = append(fields, huh.NewInput().
				Title("Go module path").
				Description("Full module path, e.g. github.com/you/my-app").
				Placeholder("github.com/you/my-app").
				Value(&modulePath).
				Validate(validateModule))
		}
		form := huh.NewForm(huh.NewGroup(fields...))
		if err := form.Run(); err != nil {
			return err
		}
	}
	if err := validateName(name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}
	if modulePath == "" {
		modulePath = name
	}
	if err := validateModule(modulePath); err != nil {
		return fmt.Errorf("invalid module path: %w", err)
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}
	if err := os.Mkdir(name, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"main.go":    strings.ReplaceAll(mainGoTemplate, "__NAME__", name),
		"index.html": strings.ReplaceAll(indexHTMLTemplate, "__NAME__", name),
		"go.mod":     strings.ReplaceAll(goModTemplate, "__MODULE__", modulePath),
	}
	for fname, content := range files {
		path := filepath.Join(name, fname)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	printTitle("Project scaffolded")
	printOK("Created %s/", styleAccent.Render(name))
	printDim("  module: %s", modulePath)
	fmt.Println()
	printInfo("Next steps:")
	printDim("  cd %s", name)
	printDim("  gutter run dev")
	fmt.Println()
	printDim("(Local checkout? Add a replace directive to %s/go.mod pointing at your gutter checkout.)", name)
	return nil
}

func validateName(s string) error {
	if s == "" {
		return fmt.Errorf("name is required")
	}
	if !nameRE.MatchString(s) {
		return fmt.Errorf("must start with a letter and contain only letters, digits, '-' or '_'")
	}
	return nil
}

func validateModule(s string) error {
	if s == "" {
		return fmt.Errorf("module path is required")
	}
	if !moduleRE.MatchString(s) {
		return fmt.Errorf("must look like a Go module path (e.g. github.com/you/project)")
	}
	return nil
}
