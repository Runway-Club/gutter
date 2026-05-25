package widgets

import (
	"github.com/Runway-Club/gutter"
)

// ImageFit picks how the image content is resized to fit its frame. The
// values map directly to CSS object-fit keywords.
type ImageFit string

const (
	// ImageFitCover scales to fill the frame, cropping overflow.
	ImageFitCover ImageFit = "cover"
	// ImageFitContain scales to fit inside the frame, letterboxing if needed.
	ImageFitContain ImageFit = "contain"
	// ImageFitFill stretches to fill the frame, distorting aspect ratio.
	ImageFitFill ImageFit = "fill"
	// ImageFitNone renders at intrinsic size, no scaling.
	ImageFitNone ImageFit = "none"
	// ImageFitScaleDown picks the smaller of None and Contain.
	ImageFitScaleDown ImageFit = "scale-down"
)

// Image renders an HTML <img>. Source can be specified two ways:
//
//   - Asset: a relative path resolved through gutter.AssetURL. The CLI build
//     copies ./assets/ into ./dist/assets/, so Asset: "logo.png" loads
//     /assets/logo.png at runtime.
//   - Src: an absolute URL (http://, https://, /abs, or data:), used as-is.
//
// If both are set, Asset wins. Width/Height accept any CSS length ("48px",
// "100%", etc.); the image's object-fit is controlled by Fit.
type Image struct {
	Asset   string
	Src     string
	Alt     string
	Width   string
	Height  string
	Fit     ImageFit
	Rounded string // optional border-radius (CSS), e.g. "50%" for an avatar
}

func (i Image) Host() *gutter.Host {
	src := i.Src
	if i.Asset != "" {
		src = gutter.AssetURL(i.Asset)
	}
	attrs := map[string]string{}
	if src != "" {
		attrs["src"] = src
	}
	if i.Alt != "" {
		attrs["alt"] = i.Alt
	}
	style := map[string]string{"display": "block"}
	if i.Width != "" {
		style["width"] = i.Width
	}
	if i.Height != "" {
		style["height"] = i.Height
	}
	if i.Fit != "" {
		style["object-fit"] = string(i.Fit)
	}
	if i.Rounded != "" {
		style["border-radius"] = i.Rounded
	}
	return &gutter.Host{
		Tag:   "img",
		Attrs: attrs,
		Style: style,
	}
}
