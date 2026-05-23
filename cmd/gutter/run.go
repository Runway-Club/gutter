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
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Build to WebAssembly and serve at :8080",
		Long:  "Build the current project to WebAssembly and serve it over HTTP. Use 'gutter run dev' for hot reload.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(addr, false)
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "address to serve on")

	dev := &cobra.Command{
		Use:   "dev",
		Short: "Build, serve, and rebuild on file changes (with live reload)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(addr, true)
		},
	}
	dev.Flags().StringVarP(&addr, "addr", "a", ":8080", "address to serve on")
	cmd.AddCommand(dev)
	return cmd
}

// buildCounter is bumped after every successful rebuild; the injected
// dev-mode client polls it and reloads when it changes.
var buildCounter atomic.Int64

func runServe(addr string, dev bool) error {
	printTitle("Initial build")
	if err := buildWasm("app.wasm"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	if err := ensureWasmExec("."); err != nil {
		return err
	}
	buildCounter.Add(1)
	printOK("app.wasm ready")
	_ = mime.AddExtensionType(".wasm", "application/wasm")

	mux := http.NewServeMux()
	if dev {
		mux.HandleFunc("/__gutter/build", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fmt.Fprintf(w, "%d", buildCounter.Load())
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Intercept index.html so we can inject the reload script.
			urlPath := strings.TrimPrefix(r.URL.Path, "/")
			if urlPath == "" || strings.HasSuffix(urlPath, "/") {
				urlPath += "index.html"
			}
			if strings.HasSuffix(urlPath, ".html") {
				data, err := os.ReadFile(filepath.Clean(urlPath))
				if err == nil {
					injected := injectReloadScript(string(data))
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.Header().Set("Cache-Control", "no-store")
					_, _ = w.Write([]byte(injected))
					return
				}
			}
			http.FileServer(http.Dir(".")).ServeHTTP(w, r)
		})
		go watchAndRebuild()
	} else {
		mux.Handle("/", http.FileServer(http.Dir(".")))
	}

	if dev {
		printTitle("Dev server")
		printInfo("watching .go files in %s and rebuilding on save", mustCWD())
	} else {
		printTitle("Server")
	}
	printOK("listening on %s", styleAccent.Render("http://localhost"+addr))
	if dev {
		printDim("(Ctrl-C to stop)")
	}
	return http.ListenAndServe(addr, mux)
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

func watchAndRebuild() {
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
			if err := buildWasm("app.wasm"); err != nil {
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
