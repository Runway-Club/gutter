//go:build !js || !wasm

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildWasm(out string, tinygo bool) error {
	return buildWasmPkg(".", out, tinygo)
}

// buildWasmPkg compiles a specific package (relative or absolute path)
// to a WASM file at out. Used by buildWasm for the app itself and by
// the worker-bundling logic for sibling Go workers. With tinygo=true it
// shells out to `tinygo build -target wasm` instead of the Go toolchain.
func buildWasmPkg(pkg, out string, tinygo bool) error {
	var cmd *exec.Cmd
	if tinygo {
		cmd = exec.Command("tinygo", "build", "-o", out, "-target", "wasm", pkg)
		cmd.Env = os.Environ()
	} else {
		cmd = exec.Command("go", "build", "-o", out, pkg)
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ensureTinygo returns a friendly error if the tinygo binary isn't on PATH,
// so `--tinygo` fails fast with install guidance rather than a cryptic
// exec error midway through the build.
func ensureTinygo() error {
	if _, err := exec.LookPath("tinygo"); err != nil {
		return fmt.Errorf("tinygo not found on PATH — install it from https://tinygo.org/getting-started/install/ (or drop --tinygo to build with the Go toolchain)")
	}
	return nil
}

// ensureWasmExec copies the appropriate wasm_exec.js into dir. It always
// overwrites: the Go and TinyGo runtimes ship different (incompatible)
// versions, so switching toolchains must replace a stale copy.
func ensureWasmExec(dir string, tinygo bool) error {
	dst := filepath.Join(dir, "wasm_exec.js")
	src, err := findWasmExec(tinygo)
	if err != nil {
		return err
	}
	return copyFile(src, dst)
}

func findWasmExec(tinygo bool) (string, error) {
	if tinygo {
		return findTinygoWasmExec()
	}
	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("locating GOROOT: %w", err)
	}
	goroot := strings.TrimSpace(string(out))
	for _, p := range []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("wasm_exec.js not found under %s (looked in lib/wasm and misc/wasm)", goroot)
}

func findTinygoWasmExec() (string, error) {
	out, err := exec.Command("tinygo", "env", "TINYGOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("locating TINYGOROOT (is tinygo installed?): %w", err)
	}
	root := strings.TrimSpace(string(out))
	p := filepath.Join(root, "targets", "wasm_exec.js")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("wasm_exec.js not found under %s", filepath.Join(root, "targets"))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
