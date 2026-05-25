package gutter

import "testing"

func TestAssetURL(t *testing.T) {
	// Restore the default base afterwards: AssetBase is package-global state.
	defer SetAssetBase("")

	cases := []struct {
		name, base, in, want string
	}{
		{"relative default base", "", "logo.png", "assets/logo.png"},
		{"relative strips leading slash off base join", "", "/logo.png", "/logo.png"}, // absolute → unchanged
		{"empty path stays empty", "", "", ""},
		{"http absolute unchanged", "", "http://cdn/x.png", "http://cdn/x.png"},
		{"https absolute unchanged", "", "https://cdn/x.png", "https://cdn/x.png"},
		{"protocol-relative unchanged", "", "//cdn/x.png", "//cdn/x.png"},
		{"root-absolute unchanged", "", "/static/x.png", "/static/x.png"},
		{"data uri unchanged", "", "data:image/svg+xml,<svg/>", "data:image/svg+xml,<svg/>"},
		{"custom base", "https://cdn.example.com/v3/", "logo.png", "https://cdn.example.com/v3/logo.png"},
		{"custom base no trailing slash gets one", "https://cdn.example.com/v3", "logo.png", "https://cdn.example.com/v3/logo.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			SetAssetBase(c.base)
			if got := AssetURL(c.in); got != c.want {
				t.Fatalf("AssetURL(%q) with base %q = %q, want %q", c.in, c.base, got, c.want)
			}
		})
	}
}

func TestSetAssetBaseNormalization(t *testing.T) {
	defer SetAssetBase("")

	SetAssetBase("cdn/assets")
	if got := AssetBaseURL(); got != "cdn/assets/" {
		t.Fatalf("AssetBaseURL() = %q, want trailing slash added", got)
	}
	SetAssetBase("") // reset
	if got := AssetBaseURL(); got != "assets/" {
		t.Fatalf("AssetBaseURL() after reset = %q, want %q", got, "assets/")
	}
}
