//go:build !js || !wasm

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// gutterICO is the Gutter brand icon, written into a scaffolded project's
// public/favicon.ico (served at /favicon.ico) and shown in the starter UI.
//
//go:embed gutter.ico
var gutterICO []byte

const mainGoTemplate = `package main

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// Root builds the app's UI. gutter.Serve calls it on the client to mount or
// hydrate, and on the server to render HTML when you run "gutter run --ssr".
func Root() gutter.Widget {
	return widgets.Scaffold{
		Title: "__NAME__",
		Theme: themes.Meta, // swap for themes.Apple or themes.Neutral
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
						Spacing:        12,
						Children: []gutter.Widget{
							// "/favicon.ico" is an absolute path, so it resolves the
							// same on every route (CSR and SSR alike).
							widgets.Image{Src: "/favicon.ico", Alt: "Gutter", Width: "72px", Height: "72px"},
							widgets.Heading{Level: widgets.H2, Text: "Hello, Gutter!"},
							// Gutter's slogan.
							widgets.Body{Text: "Build for the web in Go — the Flutter way."},
							widgets.Caption{Text: "A project by Runway Club"},
							widgets.Button{
								Variant: widgets.ButtonPrimary,
								Label:   "Get started",
							},
						},
					},
				},
			},
		},
		Footer: widgets.Center{
			Child: widgets.Caption{Text: "© Runway Club · Powered by Gutter"},
		},
	}
}

// headHTML mirrors the favicon + font <link>s from index.html, injected into
// the <head> of SSR pages ("gutter run --ssr") so they get the same icon and
// fonts as the client-rendered page. The SSR document already supplies charset,
// viewport, and a margin reset, so this only adds the brand bits.
const headHTML = "<link rel=\"icon\" type=\"image/x-icon\" href=\"/favicon.ico\">" +
	"<link rel=\"preconnect\" href=\"https://fonts.googleapis.com\">" +
	"<link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossorigin>" +
	"<link href=\"https://fonts.googleapis.com/css2?family=Lexend:wght@100..900&display=swap\" rel=\"stylesheet\">" +
	"<link href=\"https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200\" rel=\"stylesheet\">"

func main() {
	// One entry for both modes: "gutter run" serves this client-side; "gutter
	// run --ssr" builds the wasm and runs this same program as an SSR server.
	gutter.Serve(gutter.Config{Root: Root, Head: headHTML})
}
`

// ssrMainGoTemplate is the --ssr scaffold: one main.go that server-renders for
// fast first paint + SEO (gutter.Head), hydrates on the client, and demonstrates
// typed client↔server RPC. `gutter run --ssr` builds the wasm then runs this same
// program as the SSR server; `gutter run` serves it client-side.
const ssrMainGoTemplate = `package main

import (
	"context"
	"fmt"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/rpc"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// PingRequest/PingResponse are shared by the wasm client and the host server.
// Change a field and BOTH sides fail to compile — the point of typed RPC.
type PingRequest struct {
	Name string ` + "`json:\"name\"`" + `
}

type PingResponse struct {
	Message string ` + "`json:\"message\"`" + `
}

// Root is server-rendered for instant first paint + SEO (gutter.Head injects
// <title>/<meta> into the page <head>), then hydrated on the client. The same
// Root runs on both sides — gutter.Serve compiles this program twice.
func Root() gutter.Widget {
	return gutter.Head{
		Title:    "__NAME__ — built with Gutter",
		Meta:     map[string]string{"description": "A server-rendered Gutter app with typed RPC."},
		Property: map[string]string{"og:title": "__NAME__"},
		Child:    pinger{},
	}
}

type pinger struct{}

func (pinger) CreateState() gutter.State { return &pingerState{} }

type pingerState struct {
	gutter.StateObject
	reply string
	busy  bool
}

func (s *pingerState) Build(ctx *gutter.BuildContext) gutter.Widget {
	label := "Ping the server"
	if s.busy {
		label = "…"
	}
	reply := s.reply
	if reply == "" {
		reply = "(not called yet)"
	}
	return widgets.Scaffold{
		Title:  "__NAME__",
		Theme:  themes.Meta,
		AppBar: widgets.AppBar{Title: "__NAME__"},
		Body: widgets.Surface{Variant: widgets.SurfaceAlt, Child: widgets.Center{Child: widgets.Card{
			Variant: widgets.CardFeature,
			Child: widgets.Column{
				CrossAxisAlign: widgets.CrossAxisCenter,
				Spacing:        12,
				Children: []gutter.Widget{
					widgets.Image{Src: "/favicon.ico", Alt: "Gutter", Width: "72px", Height: "72px"},
					widgets.Heading{Level: widgets.H2, Text: "Server-rendered Gutter"},
					widgets.Body{Text: "First paint is HTML from the server; then wasm hydrates and this button calls Go over typed RPC."},
					widgets.Button{Variant: widgets.ButtonPrimary, Label: label, OnPressed: s.ping},
					widgets.Caption{Text: "Reply: " + reply},
				},
			},
		}}},
		Footer: widgets.Center{Child: widgets.Caption{Text: "© Runway Club · Powered by Gutter"}},
	}
}

// ping calls the server. rpc.Call blocks the goroutine on the fetch, so run it
// off the UI path and SetState the reply back in.
func (s *pingerState) ping() {
	if s.busy {
		return
	}
	s.SetState(func() { s.busy = true })
	go func() {
		res, err := rpc.Call[PingRequest, PingResponse](context.Background(), PingRequest{Name: "world"})
		s.SetState(func() {
			s.busy = false
			if err != nil {
				s.reply = "error: " + err.Error()
			} else {
				s.reply = res.Message
			}
		})
	}()
}

func main() {
	// "gutter run" serves this client-side; "gutter run --ssr" builds the wasm
	// and runs this same program as the SSR server (render + RPC + assets).
	gutter.Serve(gutter.Config{
		Root:  Root,
		Theme: themes.Meta,
		// Registered once, runs only on the server. The shared PingRequest type
		// keys the route; the client's rpc.Call reaches it with no extra wiring.
		RPC: func() {
			rpc.Handle(func(_ context.Context, r PingRequest) (PingResponse, error) {
				return PingResponse{Message: fmt.Sprintf("Hello, %s! (from the Go server)", r.Name)}, nil
			})
		},
	})
}
`

const indexHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <!-- base href="/" so relative script srcs (wasm_exec.js, app.wasm) resolve
       from the site root even when the page is loaded at a deep route like
       /user/42. Required for widgets.Router to survive page reloads. -->
  <base href="/">
  <title>__NAME__</title>
  <link rel="icon" type="image/x-icon" href="/favicon.ico">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Lexend:wght@100..900&display=swap" rel="stylesheet">
  <!-- Material Symbols (Outlined, Rounded, Sharp) for widgets.Icon. The four
       axes (FILL, wght, GRAD, opsz) are exposed so widgets.Icon can set them
       per glyph via font-variation-settings. Drop any family you don't use. -->
  <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Rounded:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Sharp:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
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

// goModTemplate is intentionally minimal — no `require` line. After scaffolding
// we run `go get github.com/Runway-Club/gutter@latest` inside the new project,
// which resolves the current published version and writes the require for us.
// If that fails (offline, network blocked, module unavailable), we leave the
// file as-is and print a hint so the user can run it themselves.
const goModTemplate = `module __MODULE__

go 1.21
`

// gitignoreTemplate ignores artifacts produced by `gutter run` / `gutter run dev`
// / `gutter build`. The CLI bundles everything into ./dist; the bare app.wasm /
// wasm_exec.js entries cover users who ran the toolchain by hand before.
const gitignoreTemplate = `# Gutter build output
/dist/
/app.wasm
/wasm_exec.js
`

var (
	nameRE   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	moduleRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)
)

func newCmd() *cobra.Command {
	var modulePath string
	var ssr bool
	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new gutter project",
		Long:  "Scaffold a new gutter project. Without arguments, prompts interactively for project name and module path. Pass --ssr for a server-rendered + typed-RPC starter.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			return runNew(name, modulePath, ssr)
		},
	}
	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (e.g. github.com/you/project)")
	cmd.Flags().BoolVar(&ssr, "ssr", false, "scaffold a server-rendered (SSR) app with typed RPC + hydration")
	return cmd
}

func runNew(name, modulePath string, ssr bool) error {
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

	mainTmpl := mainGoTemplate
	if ssr {
		mainTmpl = ssrMainGoTemplate
	}
	files := map[string]string{
		"main.go":         strings.ReplaceAll(mainTmpl, "__NAME__", name),
		"index.html":      strings.ReplaceAll(indexHTMLTemplate, "__NAME__", name),
		"go.mod":          strings.ReplaceAll(goModTemplate, "__MODULE__", modulePath),
		".gitignore":      gitignoreTemplate,
		"assets/.gitkeep": "",
	}
	for fname, content := range files {
		path := filepath.Join(name, fname)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	// The Gutter brand icon. public/ is copied to the dist root by the build, so
	// this lands at /favicon.ico — referenced by index.html and the starter UI.
	favicon := filepath.Join(name, "public", "favicon.ico")
	if err := os.MkdirAll(filepath.Dir(favicon), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(favicon), err)
	}
	if err := os.WriteFile(favicon, gutterICO, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", favicon, err)
	}

	printTitle("Project scaffolded")
	printOK("Created %s/", styleAccent.Render(name))
	printDim("  module: %s", modulePath)

	// Resolve the framework dependency to the current published version.
	// Non-fatal if it fails — the user can run the command themselves.
	gotLatest := fetchGutterLatest(name)

	fmt.Println()
	printInfo("Next steps:")
	printDim("  cd %s", name)
	if !gotLatest {
		printDim("  go get github.com/Runway-Club/gutter@latest")
	}
	if ssr {
		printDim("  gutter run --ssr   # server-rendered + typed RPC")
		printDim("  gutter run dev     # or client-side with live reload")
	} else {
		printDim("  gutter run dev")
	}
	fmt.Println()
	printDim("(Local checkout? Add a replace directive to %s/go.mod pointing at your gutter checkout.)", name)
	return nil
}

// fetchGutterLatest runs `go get github.com/Runway-Club/gutter@latest` inside
// the scaffolded project so go.mod is pinned to a real published version.
// Returns true on success; on failure logs a warning and returns false so the
// caller can print a manual instruction.
func fetchGutterLatest(projectDir string) bool {
	cmd := exec.Command("go", "get", "github.com/Runway-Club/gutter@latest")
	cmd.Dir = projectDir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		printWarn("could not resolve github.com/Runway-Club/gutter@latest: %v", err)
		if len(out) > 0 {
			printDim("  %s", strings.TrimSpace(string(out)))
		}
		return false
	}
	printOK("pinned github.com/Runway-Club/gutter@latest")
	return true
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
