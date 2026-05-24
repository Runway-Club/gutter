//go:build !js || !wasm

package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var addr string
	var tinygo bool
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Build to WebAssembly and serve at :8080",
		Long:  "Build the current project to WebAssembly and serve it over HTTP. Use 'gutter run dev' for hot reload.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(addr, false, tinygo)
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "address to serve on")
	cmd.Flags().BoolVar(&tinygo, "tinygo", false, "compile with TinyGo (much smaller .wasm; requires tinygo on PATH)")

	var devAddr string
	var devTinygo bool
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Build, serve, and rebuild on file changes (with live reload)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(devAddr, true, devTinygo)
		},
	}
	dev.Flags().StringVarP(&devAddr, "addr", "a", ":8080", "address to serve on")
	dev.Flags().BoolVar(&devTinygo, "tinygo", false, "compile with TinyGo (much smaller .wasm; requires tinygo on PATH)")
	cmd.AddCommand(dev)
	return cmd
}

// serveDir is the output directory `gutter run` / `gutter run dev` bundle into
// and serve from. Matches `gutter build`'s output, so the served bundle is
// identical to a production build (just rebuilt on every save in dev mode).
const serveDir = "dist"

// buildCounter is bumped after every successful rebuild; the injected
// dev-mode client polls it and reloads when it changes.
var buildCounter atomic.Int64

func runServe(addr string, dev, tinygo bool) error {
	printTitle("Initial build")
	if err := bundleInto(serveDir, false, tinygo); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	buildCounter.Add(1)
	printOK("bundle ready in %s", styleAccent.Render("./"+serveDir+"/"))
	_ = mime.AddExtensionType(".wasm", "application/wasm")

	mux := http.NewServeMux()
	if dev {
		mux.HandleFunc("/__gutter/build", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fmt.Fprintf(w, "%d", buildCounter.Load())
		})
	}
	mux.Handle("/", spaHandler(dev))
	if dev {
		go watchAndRebuild(tinygo)
	}

	if dev {
		printTitle("Dev server")
		printInfo("watching .go / .html / .css in %s and rebuilding on save", mustCWD())
	} else {
		printTitle("Server")
	}
	printOK("listening on %s", styleAccent.Render("http://localhost"+addr))
	if dev {
		printDim("(Ctrl-C to stop)")
	}
	return http.ListenAndServe(addr, mux)
}

// spaHandler wraps the static file server with two extras:
//
//  1. SPA fallback. If the requested path has no file extension and no file
//     matches it on disk, serve index.html instead of 404. This is what makes
//     deep links like /about or /user/42 work after a page reload — the
//     client-side Router then takes over and renders the right route.
//
//  2. Dev-mode reload script injection. When dev is true, HTML responses
//     (including the SPA fallback) get the polling snippet appended so the
//     browser reloads after every successful rebuild.
//
// File requests that DO have an extension (.wasm, .js, .css, images) skip the
// fallback and 404 normally when missing — masking those would hide real bugs.
func spaHandler(dev bool) http.Handler {
	fs := http.FileServer(http.Dir(serveDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if clean == "." {
			clean = "index.html"
		} else if strings.HasSuffix(r.URL.Path, "/") {
			clean = filepath.Join(clean, "index.html")
		}
		target := filepath.Join(serveDir, clean)

		if filepath.Ext(clean) == "" {
			if _, err := os.Stat(target); os.IsNotExist(err) {
				serveHTML(w, r, filepath.Join(serveDir, "index.html"), dev)
				return
			}
		}
		if dev && strings.HasSuffix(clean, ".html") {
			if _, err := os.Stat(target); err == nil {
				serveHTML(w, r, target, true)
				return
			}
		}
		fs.ServeHTTP(w, r)
	})
}

// serveHTML reads path, optionally injects the dev reload script, and writes
// the body. Falls back to http.NotFound if the file can't be read.
func serveHTML(w http.ResponseWriter, r *http.Request, path string, injectReload bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	body := string(data)
	if injectReload {
		body = injectReloadScript(body)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(body))
}

func mustCWD() string {
	d, err := os.Getwd()
	if err != nil {
		return "."
	}
	return d
}

const reloadSnippet = `<script>
(function () {
  let last = null;
  setInterval(async () => {
    try {
      const r = await fetch('/__gutter/build', { cache: 'no-store' });
      if (!r.ok) return;
      const v = await r.text();
      if (last === null) { last = v; return; }
      if (v !== last) { last = v; location.reload(); }
    } catch (_) {}
  }, 500);
})();
</script>`

func injectReloadScript(html string) string {
	if strings.Contains(html, "/__gutter/build") {
		return html
	}
	if i := strings.LastIndex(html, "</body>"); i >= 0 {
		return html[:i] + reloadSnippet + "\n" + html[i:]
	}
	return html + reloadSnippet
}

func watchAndRebuild(tinygo bool) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		printErr("watcher: %v", err)
		return
	}
	defer w.Close()
	if err := addGoDirs(w, "."); err != nil {
		printErr("watcher: %v", err)
		return
	}

	var (
		mu      sync.Mutex
		timer   *time.Timer
		pending bool
	)
	trigger := func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		if pending {
			return
		}
		timer = time.AfterFunc(150*time.Millisecond, func() {
			mu.Lock()
			pending = true
			mu.Unlock()

			printInfo("change detected — rebuilding")
			start := time.Now()
			if err := bundleInto(serveDir, false, tinygo); err != nil {
				printErr("rebuild failed: %v", err)
			} else {
				buildCounter.Add(1)
				printOK("rebuilt in %s — browser reloading", time.Since(start).Round(time.Millisecond))
			}

			mu.Lock()
			pending = false
			mu.Unlock()
		})
	}

	debug := os.Getenv("GUTTER_WATCH_DEBUG") != ""
	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if debug {
				printDim("watcher: %s %s", ev.Op, ev.Name)
			}
			if !shouldTrigger(ev) {
				continue
			}
			// Watch newly created directories so nested changes are seen too.
			if ev.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() && !skipDir(filepath.Base(ev.Name)) {
					_ = w.Add(ev.Name)
				}
			}
			trigger()
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			printWarn("watcher: %v", err)
		}
	}
}

func shouldTrigger(ev fsnotify.Event) bool {
	if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}
	base := filepath.Base(ev.Name)
	if strings.HasPrefix(base, ".") || strings.HasSuffix(base, "~") {
		return false
	}
	ext := filepath.Ext(base)
	switch ext {
	case ".go", ".html", ".css":
		return true
	}
	return false
}

func addGoDirs(w *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if skipDir(filepath.Base(p)) {
			return filepath.SkipDir
		}
		return w.Add(p)
	})
}

func skipDir(name string) bool {
	if name == "." || name == ".." {
		return false
	}
	switch name {
	case ".git", "node_modules", "dist", "vendor":
		return true
	}
	return strings.HasPrefix(name, ".")
}
