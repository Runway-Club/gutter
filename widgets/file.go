package widgets

import (
	"github.com/Runway-Club/gutter"
)

// FilePick is one file the user selected through a File widget. Data holds
// the raw bytes, read at pick time via FileReader on WASM (zero-length on
// the host stub).
type FilePick struct {
	Name     string
	Size     int64
	MimeType string
	Data     []byte
}

// File is a themed file picker. Clicking the trigger opens the native file
// dialog; once the user picks files they are read into memory and passed to
// OnSelect.
//
//	widgets.File{
//	    Label:    "Upload",
//	    Accept:   "image/*",
//	    Multiple: false,
//	    OnSelect: func(files []widgets.FilePick) {
//	        for _, f := range files {
//	            log.Printf("got %s (%d bytes)", f.Name, len(f.Data))
//	        }
//	    },
//	}
//
// Visually the widget renders as a Button-styled label wrapping a hidden
// <input type="file">. Variant picks the button style from the active theme.
//
// On the host (non-WASM) build the widget renders the same DOM-less label
// but OnSelect is never invoked — file reading needs the browser's
// FileReader API, which only exists under GOOS=js GOARCH=wasm.
type File struct {
	// Label is the trigger text. Ignored when Child is set.
	Label string
	// Child overrides the default label content (e.g. an Icon).
	Child gutter.Widget
	// Accept is the input's MIME filter, e.g. "image/*,application/pdf".
	Accept string
	// Multiple allows selecting more than one file.
	Multiple bool
	// OnSelect fires once all picked files have been read into memory.
	OnSelect func([]FilePick)
	// Variant selects which theme button style to use for the trigger.
	Variant ButtonVariant
}

func (f File) CreateState() gutter.State { return &fileState{} }

type fileState struct {
	gutter.StateObject
	cleanup func()
}

func (s *fileState) currentWidget() File { return s.Widget().(File) }

func (s *fileState) Build(ctx *gutter.BuildContext) gutter.Widget {
	f := s.currentWidget()
	t := activeTheme(ctx)
	style := buttonStyleFor(t, f.Variant)

	labelCSS := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border-radius":    style.Rounded,
		"padding":          style.PaddingY + " " + style.PaddingX,
		"cursor":           "pointer",
		"display":          "inline-flex",
		"align-items":      "center",
		"justify-content":  "center",
		"text-align":       "center",
		"user-select":      "none",
	}
	if style.BorderColor != "" && style.BorderWidth != "" {
		labelCSS["border"] = style.BorderWidth + " solid " + style.BorderColor
	} else {
		labelCSS["border"] = "none"
	}
	applySpec(labelCSS, style.Typography)

	inputAttrs := map[string]string{"type": "file"}
	if f.Accept != "" {
		inputAttrs["accept"] = f.Accept
	}
	if f.Multiple {
		inputAttrs["multiple"] = ""
	}

	dispatch := func(picks []FilePick) {
		cb := s.currentWidget().OnSelect
		if cb != nil {
			cb(picks)
		}
	}

	input := propSyncHost{
		tag:   "input",
		attrs: inputAttrs,
		style: map[string]string{"display": "none"},
		onMount: func(node any) {
			// OnMount fires after every reconcile update, so guard with
			// the cleanup field — we only ever wire one change listener
			// per mounted input, and Dispose tears it down.
			if s.cleanup == nil {
				s.cleanup = attachFileChangeListener(node, dispatch)
			}
		},
	}

	var triggerChild gutter.Widget
	if f.Child != nil {
		triggerChild = f.Child
	} else {
		triggerChild = Styled{Text: f.Label}
	}

	return Styled{
		Tag:   "label",
		Style: labelCSS,
		Children: []gutter.Widget{
			input,
			triggerChild,
		},
	}
}

func (s *fileState) Dispose() {
	if s.cleanup != nil {
		s.cleanup()
		s.cleanup = nil
	}
}
