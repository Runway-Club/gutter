package gutter

import (
	"strings"
	"sync"
)

// AssetBase is the URL prefix prepended to every asset path resolved through
// AssetURL. It is "assets/" by default — the conventional directory the
// `gutter build` CLI copies from ./assets/ into ./dist/assets/. Override it
// at startup (e.g. to point at a CDN) before any widget renders an asset:
//
//	gutter.SetAssetBase("https://cdn.example.com/v3/")
//	gutter.RunApp(App{})
//
// AssetBase always ends with "/" after SetAssetBase normalizes it.
var (
	assetMu   sync.RWMutex
	assetBase = "assets/"
)

// SetAssetBase replaces the asset base URL. A trailing slash is added if
// missing. Pass "" to reset to the default ("assets/").
func SetAssetBase(base string) {
	assetMu.Lock()
	defer assetMu.Unlock()
	if base == "" {
		assetBase = "assets/"
		return
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	assetBase = base
}

// AssetBaseURL returns the current asset base URL.
func AssetBaseURL() string {
	assetMu.RLock()
	defer assetMu.RUnlock()
	return assetBase
}

// AssetURL resolves a relative asset path against the configured base URL.
// A path that already looks absolute ("http://...", "https://...", "//cdn",
// "/abs", or "data:...") is returned unchanged so widgets can accept either
// declared assets or arbitrary URLs through the same field.
//
//	gutter.AssetURL("logo.png")                  // "assets/logo.png"
//	gutter.AssetURL("/static/logo.png")           // "/static/logo.png"
//	gutter.AssetURL("https://cdn/logo.png")       // "https://cdn/logo.png"
func AssetURL(path string) string {
	if path == "" {
		return ""
	}
	if isAbsoluteURL(path) {
		return path
	}
	base := AssetBaseURL()
	return base + strings.TrimPrefix(path, "/")
}

func isAbsoluteURL(s string) bool {
	if strings.HasPrefix(s, "/") {
		return true
	}
	if strings.HasPrefix(s, "data:") {
		return true
	}
	if i := strings.Index(s, "://"); i > 0 {
		return true
	}
	return false
}
