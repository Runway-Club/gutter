// Showcase: a tour of every widget in the gutter catalog under whichever
// theme is chosen at build time. Defaults to Apple. To render with Meta:
//
//	cd examples/showcase
//	GOOS=js GOARCH=wasm go build -ldflags "-X 'main.themeName=meta'" -o app.wasm .
//
// The page is one long scrollable column. Sections demonstrate primitives,
// the input family, overlays, the imperative/canvas/worker escape hatches,
// the reactive plumbing (Notifier+ObserverBuilder, AsyncBuilder,
// AnimationController, Router), and the List/ListBuilder pair — including
// a 10k-row virtualized list to prove the recycling actually scales.
//
// The "Continue with Google" button is wired through
// github.com/Runway-Club/gutter/community/login_with_google — a reusable
// package that lives next to the core widget catalog.
package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Runway-Club/gutter"
	loginwithgoogle "github.com/Runway-Club/gutter/community/login_with_google"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// themeName is set via -ldflags. Defaults to apple.
var themeName = "meta"

// googleClientID is the OAuth 2.0 Client ID this showcase uses for the
// LoginWithGoogle button. App-specific — kept in the showcase, not in the
// reusable community package. Replace with your own ID after adding the
// origin to "Authorized JavaScript origins" in Google Cloud Console.
const googleClientID = "836815761802-cpm6e1s7psn4b1s9vursci1q4psop8p9.apps.googleusercontent.com"

// reverseTask is the Web Worker payload. Registering the WorkerTask at
// package scope means the worker bootstrap (which reloads app.wasm with
// __GUTTER_WORKER_TASK="reverse") can find it before RunApp runs.
var reverseTask = gutter.NewWorkerTask("reverse", func(msg string) string {
	runes := []rune(msg)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
})

// Showcase is a StatefulWidget — every interactive section reads/writes
// a field on the single owning State.
type Showcase struct{}

func (Showcase) CreateState() gutter.State { return &showcaseState{} }

type showcaseState struct {
	gutter.StateObject

	// Form/input state.
	textValue     string
	emailValue    string
	passwordValue string
	numberValue   string
	colorValue    string
	textAreaValue string
	checkboxOn    bool
	switchOn      bool
	sliderValue   float64
	selectColor   string
	radioSize     string
	pickedFiles   []widgets.FilePick
	asyncReloadID int

	// Gesture demo.
	gestureX, gestureY float64
	gestureTaps        int

	// Google sign-in result.
	googleUser *loginwithgoogle.Credential
	googleErr  string

	// Cross-tree state — driven through Notifiers so ObserverBuilder /
	// overlay widgets can subscribe directly without going through SetState.
	counter    *gutter.Notifier[int]
	popupOpen  *gutter.Notifier[bool]
	drawerOpen *gutter.Notifier[bool]
	sheetOpen  *gutter.Notifier[bool]

	anim       *widgets.AnimationController
	miniRouter *widgets.Router
}

func (s *showcaseState) InitState() {
	s.counter = gutter.NewNotifier(0)
	s.popupOpen = gutter.NewNotifier(false)
	s.drawerOpen = gutter.NewNotifier(false)
	s.sheetOpen = gutter.NewNotifier(false)

	s.textValue = "Hello"
	s.colorValue = "#0066cc"
	s.sliderValue = 42
	s.selectColor = "blue"
	s.radioSize = "m"

	s.anim = widgets.NewAnimationController(900 * time.Millisecond)
	s.anim.Curve = widgets.CurveEaseInOut

	routes := map[string]widgets.RouteBuilder{
		"/": func(_ widgets.RouteParams) gutter.Widget { return miniRoutePane("Home", "Top-level route for /") },
		"/specs": func(_ widgets.RouteParams) gutter.Widget {
			return miniRoutePane("Specs", "Inner route /specs — try the browser back button.")
		},
		"/help": func(_ widgets.RouteParams) gutter.Widget {
			return miniRoutePane("Help", "Inner route /help — same Router instance.")
		},
	}
	s.miniRouter = widgets.NewRouter(routes, miniRoutePane("?", "no route"))
	// The page can be opened at "/index.html" (most static servers) which
	// doesn't match any inner route. Normalize to "/" so first paint shows
	// the Home pane instead of the not-found fallback.
	if _, ok := routes[s.miniRouter.Current()]; !ok {
		s.miniRouter.Replace("/")
	}
}

func (s *showcaseState) Dispose() {
	if s.anim != nil {
		s.anim.Stop()
	}
}

func (s *showcaseState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title:        "Gutter Showcase",
		Theme:        pickTheme(),
		StickyAppBar: true,
		AppBar: widgets.AppBar{
			Title:   "Gutter",
			Leading: widgets.IconButton{Icon: "menu", Variant: widgets.ButtonGhost, OnPressed: func() { s.drawerOpen.Set(true) }, Tooltip: "Open drawer"},
			Actions: []gutter.Widget{
				widgets.IconButton{Icon: "search", Variant: widgets.ButtonGhost, Tooltip: "Search"},
				widgets.IconButton{Icon: "settings", Variant: widgets.ButtonGhost, OnPressed: func() { s.sheetOpen.Set(true) }, Tooltip: "Settings"},
			},
		},
		Body: widgets.Column{
			Children: []gutter.Widget{
				heroSection(ctx),
				typographySection(),
				buttonsSection(),
				iconsSection(),
				cardsSection(),
				surfaceVariantsSection(),
				badgesSection(),
				imagesSection(),
				layoutSection(),
				inputsSection(s),
				formControlsSection(s),
				textAreaSection(s),
				fileSection(s),
				gestureSection(s),
				animationSection(s),
				canvasSection(),
				observerSection(s),
				asyncSection(s),
				workerSection(),
				listSection(),
				listBuilderSection(),
				routerSection(s),
				authSection(s),
				footerSection(),
				// Overlays. They render position:fixed so they don't
				// participate in the column layout; siblings at any
				// level work.
				widgets.Popup{
					Open:      s.popupOpen,
					OnDismiss: func() { s.popupOpen.Set(false) },
					Child: widgets.Column{
						Spacing: 12,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H4, Text: "Hello from a Popup"},
							widgets.Body{Text: "Click the backdrop or the button to dismiss."},
							widgets.Button{Label: "Close", OnPressed: func() { s.popupOpen.Set(false) }},
						},
					},
				},
				widgets.Drawer{
					Open:      s.drawerOpen,
					Side:      widgets.DrawerLeft,
					OnDismiss: func() { s.drawerOpen.Set(false) },
					Child: widgets.Column{
						Spacing: 16,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H4, Text: "Drawer"},
							widgets.Link{Text: "Typography", OnPressed: func() { s.drawerOpen.Set(false) }},
							widgets.Link{Text: "Inputs", OnPressed: func() { s.drawerOpen.Set(false) }},
							widgets.Link{Text: "Overlays", OnPressed: func() { s.drawerOpen.Set(false) }},
						},
					},
				},
				widgets.BottomSheet{
					Open:      s.sheetOpen,
					OnDismiss: func() { s.sheetOpen.Set(false) },
					Child: widgets.Column{
						Spacing: 8,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H4, Text: "Quick actions"},
							widgets.Button{Variant: widgets.ButtonGhost, Label: "Share"},
							widgets.Button{Variant: widgets.ButtonGhost, Label: "Duplicate"},
							widgets.Button{Variant: widgets.ButtonGhost, Label: "Archive"},
						},
					},
				},
			},
		},
	}
}

// ============================ sections ============================

func sectionFrame(title string, child gutter.Widget) gutter.Widget {
	return widgets.Surface{
		Variant: widgets.SurfaceCanvas,
		Padding: "48px 32px",
		Child: widgets.Column{
			Spacing: 16,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H4, Text: title},
				child,
			},
		},
	}
}

func heroSection(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Surface{
		Variant: widgets.SurfaceDark,
		Child: widgets.Center{
			Child: widgets.Column{
				CrossAxisAlign: widgets.CrossAxisCenter,
				Spacing:        24,
				Children: []gutter.Widget{
					widgets.Heading{Level: widgets.H1, Text: "Every widget. One page.", Color: ctx.Theme.Colors.OnDark},
					widgets.Body{Text: "Scroll through the catalog — primitives, inputs, overlays, animations, lists.", Color: ctx.Theme.Colors.OnDark},
					widgets.Row{
						Spacing: 12,
						Children: []gutter.Widget{
							widgets.Button{Variant: widgets.ButtonOnDark, Label: "Get started"},
							widgets.Button{Variant: widgets.ButtonGhost, Label: "Read the docs"},
						},
					},
				},
			},
		},
	}
}

func typographySection() gutter.Widget {
	return sectionFrame("Typography", widgets.Column{
		Spacing: 6,
		Children: []gutter.Widget{
			widgets.Heading{Level: widgets.H1, Text: "H1 hero display"},
			widgets.Heading{Level: widgets.H2, Text: "H2 display large"},
			widgets.Heading{Level: widgets.H3, Text: "H3 display medium"},
			widgets.Heading{Level: widgets.H4, Text: "H4 heading large"},
			widgets.Heading{Level: widgets.H5, Text: "H5 heading medium"},
			widgets.Heading{Level: widgets.H6, Text: "H6 heading small"},
			widgets.Body{Text: "Body — comfortable reading paragraph text."},
			widgets.Body{Text: "Body / bold variant.", Bold: true},
			widgets.Body{Text: "Body / small variant — drops to caption size.", Small: true},
			widgets.Caption{Text: "Caption — for incidental UI metadata."},
			widgets.Caption{Text: "Caption / strong", Bold: true},
			widgets.Link{Text: "Inline link with a click handler", OnPressed: func() {}},
			// Text + TextStyle: the raw primitive.
			widgets.Text{Data: "widgets.Text with manual TextStyle", Style: &widgets.TextStyle{Color: "#888", FontSize: "14px", FontWeight: "300"}},
		},
	})
}

func buttonsSection() gutter.Widget {
	return sectionFrame("Buttons", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Row{
				Spacing: 12,
				Children: []gutter.Widget{
					widgets.Button{Variant: widgets.ButtonPrimary, Label: "Primary"},
					widgets.Button{Variant: widgets.ButtonSecondary, Label: "Secondary"},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Ghost"},
					widgets.Button{Variant: widgets.ButtonAccent, Label: "Accent"},
				},
			},
			widgets.Row{
				Spacing:        12,
				CrossAxisAlign: widgets.CrossAxisCenter,
				Children: []gutter.Widget{
					widgets.IconButton{Icon: "favorite", Variant: widgets.ButtonGhost, Tooltip: "Like", Filled: true},
					widgets.IconButton{Icon: "share", Variant: widgets.ButtonGhost, Tooltip: "Share"},
					widgets.IconButton{Icon: "delete", Variant: widgets.ButtonGhost, Tooltip: "Delete", IconStyle: widgets.IconRounded},
					widgets.IconButton{Icon: "bookmark", Variant: widgets.ButtonPrimary, Tooltip: "Bookmark"},
				},
			},
		},
	})
}

func iconsSection() gutter.Widget {
	icons := []string{"home", "settings", "favorite", "search", "delete", "info", "warning", "check_circle", "calendar_today", "mail"}
	row := make([]gutter.Widget, 0, len(icons))
	for _, name := range icons {
		row = append(row, widgets.Column{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        4,
			Children: []gutter.Widget{
				widgets.Icon{Name: name, Size: "32px"},
				widgets.Caption{Text: name},
			},
		})
	}
	row = append(row,
		widgets.Column{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        4,
			Children: []gutter.Widget{
				widgets.Icon{Name: "favorite", Size: "32px", Style: widgets.IconRounded, Filled: true, Color: "#e91e63"},
				widgets.Caption{Text: "rounded/filled"},
			},
		},
		widgets.Column{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        4,
			Children: []gutter.Widget{
				widgets.Icon{Name: "star", Size: "32px", Style: widgets.IconSharp, Weight: 700},
				widgets.Caption{Text: "sharp/700"},
			},
		},
	)
	return sectionFrame("Icons (Material Symbols)", widgets.Row{Spacing: 24, Children: row})
}

func cardsSection() gutter.Widget {
	return sectionFrame("Cards", widgets.Row{
		Spacing: 16,
		Children: []gutter.Widget{
			widgets.Card{Variant: widgets.CardFeature, Child: widgets.Column{Spacing: 8, Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H5, Text: "Feature"},
				widgets.Body{Text: "A bordered light card.", Small: true},
			}}},
			widgets.Card{Variant: widgets.CardPlain, Child: widgets.Column{Spacing: 8, Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H5, Text: "Plain"},
				widgets.Body{Text: "Minimally-decorated rounded surface.", Small: true},
			}}},
		},
	})
}

func surfaceVariantsSection() gutter.Widget {
	// Three Surface variants stacked so each renders its actual background.
	tile := func(variant widgets.SurfaceVariant, label string) gutter.Widget {
		return widgets.Container{
			Width:  "100%",
			Height: "120px",
			Child: widgets.Surface{
				Variant: variant,
				Padding: "16px",
				Child:   widgets.Heading{Level: widgets.H6, Text: label},
			},
		}
	}
	return sectionFrame("Surface variants", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			tile(widgets.SurfaceCanvas, "Canvas"),
			tile(widgets.SurfaceAlt, "Alt"),
			tile(widgets.SurfaceDark, "Dark"),
		},
	})
}

func badgesSection() gutter.Widget {
	return sectionFrame("Badges", widgets.Row{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Badge{Variant: widgets.BadgeNeutral, Text: "Neutral"},
			widgets.Badge{Variant: widgets.BadgeSuccess, Text: "In stock"},
			widgets.Badge{Variant: widgets.BadgeWarning, Text: "Selling fast"},
			widgets.Badge{Variant: widgets.BadgeCritical, Text: "Out of stock"},
		},
	})
}

func imagesSection() gutter.Widget {
	return sectionFrame("Image", widgets.Row{
		Spacing:        24,
		CrossAxisAlign: widgets.CrossAxisCenter,
		Children: []gutter.Widget{
			widgets.Column{
				Spacing:        4,
				CrossAxisAlign: widgets.CrossAxisCenter,
				Children: []gutter.Widget{
					widgets.Image{Src: loginwithgoogle.GLogoDataURL, Width: "64px", Height: "64px"},
					widgets.Caption{Text: "Inline SVG data URL"},
				},
			},
			widgets.Column{
				Spacing:        4,
				CrossAxisAlign: widgets.CrossAxisCenter,
				Children: []gutter.Widget{
					widgets.Image{Src: loginwithgoogle.GLogoDataURL, Width: "64px", Height: "64px", Rounded: "50%"},
					widgets.Caption{Text: "Rounded 50%"},
				},
			},
			widgets.Column{
				Spacing:        4,
				CrossAxisAlign: widgets.CrossAxisCenter,
				Children: []gutter.Widget{
					widgets.Image{Asset: "sample.svg", Width: "64px", Height: "64px", Fit: widgets.ImageFitContain},
					widgets.Caption{Text: "Asset path"},
				},
			},
		},
	})
}

// layoutSection tours the flex/grid layout primitives and the theme Color
// tokens. Every coloured box here is a Container tinted with a token
// (ColorSurfaceSoft, ColorSurfaceDark, …) rather than a hard-coded hex — swap
// the build-time theme and the whole section recolours itself.
func layoutSection() gutter.Widget {
	tile := func(label string) gutter.Widget {
		return widgets.Container{
			Color:        widgets.ColorSurfaceSoft,
			BorderColor:  widgets.ColorHairline,
			BorderRadius: "10px",
			Padding:      widgets.EdgeInsetsAll(14),
			Child:        widgets.Body{Text: label, Small: true},
		}
	}
	gridTiles := make([]gutter.Widget, 6)
	for i := range gridTiles {
		gridTiles[i] = tile(fmt.Sprintf("cell %d", i+1))
	}
	chipLabels := []string{"flutter", "wasm", "go", "declarative", "no-css", "reactive", "themed"}
	chips := make([]gutter.Widget, len(chipLabels))
	for i, c := range chipLabels {
		chips[i] = widgets.Container{
			Color:        widgets.ColorCanvasAlt,
			BorderColor:  widgets.ColorHairline,
			BorderRadius: "999px",
			Padding:      widgets.EdgeInsetsSymmetric(6, 14),
			Child:        widgets.Caption{Text: c},
		}
	}
	tokens := []struct{ name, token string }{
		{"Primary", widgets.ColorPrimary},
		{"Accent", widgets.ColorAccent},
		{"SurfaceSoft", widgets.ColorSurfaceSoft},
		{"CanvasAlt", widgets.ColorCanvasAlt},
		{"SurfaceDark", widgets.ColorSurfaceDark},
		{"Success", widgets.ColorSuccess},
		{"Warning", widgets.ColorWarning},
		{"Critical", widgets.ColorCritical},
	}
	swatches := make([]gutter.Widget, len(tokens))
	for i, tk := range tokens {
		swatches[i] = widgets.Column{
			Spacing:        4,
			CrossAxisAlign: widgets.CrossAxisCenter,
			Children: []gutter.Widget{
				widgets.Container{Color: tk.token, BorderRadius: "8px", Width: "72px", Height: "44px", BorderColor: widgets.ColorHairline},
				widgets.Caption{Text: tk.name},
			},
		}
	}

	return sectionFrame("Layout & color tokens", widgets.Column{
		Spacing: 24,
		Children: []gutter.Widget{
			// Row + Expanded + Spacer: a fixed label, an Expanded that eats the
			// free space, then an action pushed to the far right by a Spacer.
			widgets.Caption{Text: "Row · Expanded · Spacer"},
			widgets.Row{
				Spacing:        8,
				CrossAxisAlign: widgets.CrossAxisCenter,
				Children: []gutter.Widget{
					tile("fixed"),
					widgets.Expanded{Child: tile("Expanded — fills the remaining width")},
					widgets.Spacer{},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Action"},
				},
			},

			// Stack + Positioned: a card with a badge pinned to its corner.
			widgets.Caption{Text: "Stack · Positioned (corner badge)"},
			widgets.Stack{
				Width:  "180px",
				Height: "104px",
				Children: []gutter.Widget{
					widgets.Container{
						Color:        widgets.ColorCanvasAlt,
						BorderColor:  widgets.ColorHairline,
						BorderRadius: "12px",
						Width:        "180px",
						Height:       "104px",
						Child:        widgets.Center{Child: widgets.Body{Text: "base layer"}},
					},
					widgets.Positioned{Top: "-8px", Right: "-8px", Child: widgets.Badge{Text: "NEW"}},
				},
			},

			// Responsive grid — no media query, just minmax(auto-fill).
			widgets.Caption{Text: "Grid · MinColumnWidth 140px (resize the window — it reflows)"},
			widgets.Grid{MinColumnWidth: "140px", Gap: 12, Children: gridTiles},

			// Wrap: chips that flow onto new lines.
			widgets.Caption{Text: "Wrap · chips"},
			widgets.Wrap{Spacing: 8, RunSpacing: 8, Children: chips},

			// AspectRatio inside a ConstrainedBox.
			widgets.Caption{Text: "ConstrainedBox MaxWidth 320 · AspectRatio 16:9"},
			widgets.ConstrainedBox{
				MaxWidth: "320px",
				Child: widgets.AspectRatio{
					Ratio: 16.0 / 9.0,
					Child: widgets.Container{
						Color:        widgets.ColorSurfaceDark,
						BorderRadius: "12px",
						Child:        widgets.Center{Child: widgets.Body{Text: "16 : 9", Color: widgets.ColorOnDark}},
					},
				},
			},

			// Align: child anchored bottom-right of a fixed box.
			widgets.Caption{Text: "Align · BottomRight"},
			widgets.Container{
				Color:        widgets.ColorSurfaceSoft,
				BorderColor:  widgets.ColorHairline,
				BorderRadius: "12px",
				Width:        "100%",
				Height:       "96px",
				Child:        widgets.Align{Alignment: widgets.AlignBottomRight, Child: widgets.Padding{Padding: widgets.EdgeInsetsAll(10), Child: widgets.Badge{Text: "pinned"}}},
			},

			// The palette, every box tinted by a token.
			widgets.Caption{Text: "Color tokens — resolved against the active theme"},
			widgets.Wrap{Spacing: 16, RunSpacing: 12, Children: swatches},
		},
	})
}

func inputsSection(s *showcaseState) gutter.Widget {
	return sectionFrame("Inputs (Type variants)", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Input{Type: widgets.InputText, Placeholder: "Plain text", Value: s.textValue, OnChanged: func(v string) { s.SetState(func() { s.textValue = v }) }},
			widgets.Input{Type: widgets.InputEmail, Placeholder: "you@example.com", Value: s.emailValue, OnChanged: func(v string) { s.SetState(func() { s.emailValue = v }) }, AutoComplete: "email"},
			widgets.Input{Type: widgets.InputPassword, Placeholder: "Password", Value: s.passwordValue, OnChanged: func(v string) { s.SetState(func() { s.passwordValue = v }) }, AutoComplete: "current-password"},
			widgets.Input{Type: widgets.InputNumber, Placeholder: "0", Value: s.numberValue, OnChanged: func(v string) { s.SetState(func() { s.numberValue = v }) }, Min: "0", Max: "100", Step: "1"},
			widgets.Input{Type: widgets.InputSearch, Placeholder: "Search…"},
			widgets.Input{Type: widgets.InputTel, Placeholder: "+1 555 555 5555"},
			widgets.Input{Type: widgets.InputURL, Placeholder: "https://"},
			widgets.Input{Type: widgets.InputDate},
			widgets.Input{Type: widgets.InputTime},
			widgets.Input{Type: widgets.InputDateTimeLocal},
			widgets.Input{Type: widgets.InputMonth},
			widgets.Input{Type: widgets.InputWeek},
			widgets.Input{Type: widgets.InputColor, Value: s.colorValue, OnChanged: func(v string) { s.SetState(func() { s.colorValue = v }) }},
			widgets.Input{Type: widgets.InputText, Placeholder: "Disabled", Disabled: true},
			widgets.Input{Type: widgets.InputText, Placeholder: "Read-only", Value: "set externally", ReadOnly: true},
			widgets.Input{Type: widgets.InputText, Placeholder: "Error state", Error: true},
		},
	})
}

func formControlsSection(s *showcaseState) gutter.Widget {
	return sectionFrame("Form controls", widgets.Column{
		Spacing: 16,
		Children: []gutter.Widget{
			widgets.Checkbox{
				Checked:   s.checkboxOn,
				Label:     "I agree to the terms",
				OnChanged: func(v bool) { s.SetState(func() { s.checkboxOn = v }) },
			},
			widgets.Switch{
				Checked:   s.switchOn,
				Label:     "Send weekly digest",
				OnChanged: func(v bool) { s.SetState(func() { s.switchOn = v }) },
			},
			widgets.Column{Spacing: 4, Children: []gutter.Widget{
				widgets.Caption{Text: fmt.Sprintf("Slider value: %.0f", s.sliderValue)},
				widgets.Slider{
					Value:     s.sliderValue,
					Min:       0,
					Max:       100,
					Step:      1,
					OnChanged: func(v float64) { s.SetState(func() { s.sliderValue = v }) },
				},
			}},
			widgets.Select[string]{
				Placeholder: "Pick a color",
				Options: []widgets.SelectOption[string]{
					{Value: "red", Label: "Red"},
					{Value: "green", Label: "Green"},
					{Value: "blue", Label: "Blue"},
				},
				Selected:  s.selectColor,
				OnChanged: func(v string) { s.SetState(func() { s.selectColor = v }) },
			},
			widgets.RadioGroup[string]{
				Direction: "row",
				Options: []widgets.RadioOption[string]{
					{Value: "s", Label: "Small"},
					{Value: "m", Label: "Medium"},
					{Value: "l", Label: "Large"},
				},
				Selected:  s.radioSize,
				OnChanged: func(v string) { s.SetState(func() { s.radioSize = v }) },
			},
			widgets.Caption{Text: fmt.Sprintf("Picks: checkbox=%v, switch=%v, color=%s, size=%s", s.checkboxOn, s.switchOn, s.selectColor, s.radioSize)},
		},
	})
}

func textAreaSection(s *showcaseState) gutter.Widget {
	return sectionFrame("TextArea", widgets.Column{
		Spacing: 8,
		Children: []gutter.Widget{
			widgets.TextArea{
				Placeholder: "Long-form text…",
				Value:       s.textAreaValue,
				Rows:        5,
				MaxLength:   500,
				OnChanged:   func(v string) { s.SetState(func() { s.textAreaValue = v }) },
			},
			widgets.Caption{Text: fmt.Sprintf("%d / 500 characters", len(s.textAreaValue))},
		},
	})
}

func fileSection(s *showcaseState) gutter.Widget {
	picked := "nothing picked yet"
	if len(s.pickedFiles) > 0 {
		picked = fmt.Sprintf("%d file(s): first is %q (%d bytes, %s)",
			len(s.pickedFiles), s.pickedFiles[0].Name, len(s.pickedFiles[0].Data), s.pickedFiles[0].MimeType)
	}
	return sectionFrame("File picker", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.File{
				Label:    "Choose files",
				Accept:   "image/*,application/pdf",
				Multiple: true,
				OnSelect: func(files []widgets.FilePick) {
					s.SetState(func() { s.pickedFiles = files })
				},
			},
			widgets.Caption{Text: picked},
		},
	})
}

func gestureSection(s *showcaseState) gutter.Widget {
	hit := widgets.GestureDetector{
		OnTap: func() { s.SetState(func() { s.gestureTaps++ }) },
		OnPointerMove: func(e gutter.Event) {
			s.SetState(func() {
				s.gestureX = e.OffsetX
				s.gestureY = e.OffsetY
			})
		},
		Child: widgets.Container{
			Width:        "100%",
			Height:       "120px",
			Color:        "#0066cc20",
			BorderRadius: "12px",
			Child: widgets.Center{Child: widgets.Caption{
				Text: fmt.Sprintf("Taps: %d  •  Pointer: (%.0f, %.0f)", s.gestureTaps, s.gestureX, s.gestureY),
			}},
		},
	}
	return sectionFrame("GestureDetector", hit)
}

func animationSection(s *showcaseState) gutter.Widget {
	return sectionFrame("Animation + Transform", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Row{
				Spacing: 12,
				Children: []gutter.Widget{
					widgets.Button{Variant: widgets.ButtonPrimary, Label: "Forward", OnPressed: func() { s.anim.Forward() }},
					widgets.Button{Variant: widgets.ButtonSecondary, Label: "Reverse", OnPressed: func() { s.anim.Reverse() }},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Reset", OnPressed: func() { s.anim.Reset() }},
				},
			},
			widgets.AnimatedBuilder{
				Controller: s.anim,
				Builder: func(_ *gutter.BuildContext, t float64) gutter.Widget {
					return widgets.Padding{
						Padding: widgets.EdgeInsetsSymmetric(16, 0),
						Child: widgets.Transform{
							TranslateX: 240 * t,
							Rotate:     360 * t,
							Scale:      0.5 + 0.5*t,
							Child: widgets.Container{
								Width:        "64px",
								Height:       "64px",
								Color:        "#0066cc",
								BorderRadius: "12px",
							},
						},
					}
				},
			},
		},
	})
}

func canvasSection() gutter.Widget {
	bars := []float64{0.3, 0.7, 0.5, 0.9, 0.4, 0.65, 0.85}
	return sectionFrame("Canvas painter", widgets.Canvas{
		Width:      480,
		Height:     160,
		Background: "#f5f5f7",
		Paint: func(p *widgets.CanvasPainter) {
			w, h := p.Size()
			p.Clear()
			barW := w / float64(len(bars)*2)
			for i, v := range bars {
				x := float64(i)*barW*2 + barW*0.5
				barH := v * (h - 32)
				p.FillStyle(fmt.Sprintf("rgb(%d,102,204)", 30+i*24))
				p.FillRect(x, h-barH-8, barW, barH)
			}
			p.FillStyle("#666")
			p.Font("14px Lexend")
			p.TextBaseline("top")
			p.FillText("painted via syscall/js", 8, 8)
		},
	})
}

func observerSection(s *showcaseState) gutter.Widget {
	return sectionFrame("Notifier + ObserverBuilder", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.ObserverBuilder[int]{
				Source: s.counter,
				Builder: func(_ *gutter.BuildContext, v int) gutter.Widget {
					return widgets.Heading{Level: widgets.H3, Text: fmt.Sprintf("Counter: %d", v)}
				},
			},
			widgets.Row{Spacing: 8, Children: []gutter.Widget{
				widgets.IconButton{Icon: "remove", Variant: widgets.ButtonGhost, OnPressed: func() { s.counter.Update(func(v int) int { return v - 1 }) }},
				widgets.IconButton{Icon: "add", Variant: widgets.ButtonPrimary, OnPressed: func() { s.counter.Update(func(v int) int { return v + 1 }) }},
				widgets.Button{Variant: widgets.ButtonSecondary, Label: "Reset", OnPressed: func() { s.counter.Set(0) }},
			}},
		},
	})
}

func asyncSection(s *showcaseState) gutter.Widget {
	return sectionFrame("AsyncBuilder", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Button{
				Variant:   widgets.ButtonSecondary,
				Label:     "Refetch",
				OnPressed: func() { s.SetState(func() { s.asyncReloadID++ }) },
			},
			widgets.WithKey{
				Key: s.asyncReloadID,
				Child: widgets.AsyncBuilder[string]{
					Load: func(ctx context.Context) (string, error) {
						select {
						case <-ctx.Done():
							return "", ctx.Err()
						case <-time.After(700 * time.Millisecond):
							return fmt.Sprintf("loaded at %s", time.Now().Format(time.Kitchen)), nil
						}
					},
					Builder: func(_ *gutter.BuildContext, snap widgets.AsyncSnapshot[string]) gutter.Widget {
						switch snap.State {
						case widgets.AsyncPending:
							return widgets.Body{Text: "Loading…"}
						case widgets.AsyncFailed:
							return widgets.Body{Text: "Error: " + snap.Error.Error()}
						}
						return widgets.Body{Text: snap.Data}
					},
				},
			},
		},
	})
}

func workerSection() gutter.Widget {
	return sectionFrame("Worker (inline task)", WorkerDemo{})
}

// WorkerDemo wraps a Worker widget with an Input feeding it on demand.
// Lives here in main.go rather than the widgets package because it's
// example glue — a stateful little harness that holds the input string
// and uses snap.Post to send it on every keystroke (debounced trivially
// by the Worker itself: each Post supersedes the previous).
type WorkerDemo struct{}

func (WorkerDemo) CreateState() gutter.State { return &workerDemoState{} }

type workerDemoState struct {
	gutter.StateObject
	text string
}

func (s *workerDemoState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Column{
		Spacing: 8,
		Children: []gutter.Widget{
			widgets.Input{
				Type:        widgets.InputText,
				Placeholder: "Type something — the worker reverses it",
				Value:       s.text,
				OnChanged:   func(v string) { s.SetState(func() { s.text = v }) },
			},
			widgets.Worker{
				Task:    reverseTask,
				Message: s.text,
				Builder: func(snap widgets.WorkerSnapshot) gutter.Widget {
					if snap.Pending {
						return widgets.Body{Text: "(reversing…)", Small: true}
					}
					if snap.Error != "" {
						return widgets.Body{Text: "Error: " + snap.Error, Small: true}
					}
					if snap.Message == "" {
						return widgets.Body{Text: "(awaiting input)", Small: true}
					}
					return widgets.Body{Text: "Reversed: " + snap.Message, Bold: true}
				},
			},
		},
	}
}

func listSection() gutter.Widget {
	items := []gutter.Widget{}
	for i := 0; i < 8; i++ {
		items = append(items, widgets.Container{
			Padding:      widgets.EdgeInsetsAll(12),
			BorderRadius: "8px",
			Color:        "#fafafc",
			Child:        widgets.Body{Text: fmt.Sprintf("List row %d", i+1)},
		})
	}
	return sectionFrame("List (eager scroll)", widgets.List{
		Children: items,
		Spacing:  8,
		Height:   "240px",
		Padding:  widgets.EdgeInsetsAll(8),
	})
}

func listBuilderSection() gutter.Widget {
	return sectionFrame("ListBuilder (virtualized, 10,000 rows)", widgets.ListBuilder{
		ItemCount:  10000,
		ItemHeight: 56,
		Height:     "360px",
		ItemBuilder: func(i int) gutter.Widget {
			// Same widget type for every row — that's the recycling
			// contract: positional matching across scroll lets the
			// reconciler update DOM in place.
			return widgets.Container{
				Padding:      widgets.EdgeInsetsSymmetric(8, 16),
				BorderRadius: "8px",
				Color: func() string {
					if i%2 == 0 {
						return "#ffffff"
					}
					return "#f5f5f7"
				}(),
				Child: widgets.Row{
					CrossAxisAlign: widgets.CrossAxisCenter,
					Spacing:        12,
					Children: []gutter.Widget{
						widgets.Icon{Name: "drag_indicator", Size: "20px", Color: "#999"},
						widgets.Body{Text: fmt.Sprintf("Row %05d — virtualized", i)},
						widgets.SizedBox{Width: "16px"},
						widgets.Badge{Variant: badgeFor(i), Text: fmt.Sprintf("#%d", i)},
					},
				},
			}
		},
	})
}

func badgeFor(i int) widgets.BadgeVariant {
	switch i % 4 {
	case 0:
		return widgets.BadgeNeutral
	case 1:
		return widgets.BadgeSuccess
	case 2:
		return widgets.BadgeWarning
	default:
		return widgets.BadgeCritical
	}
}

func routerSection(s *showcaseState) gutter.Widget {
	return sectionFrame("Router + RouterView", widgets.Column{
		Spacing: 12,
		Children: []gutter.Widget{
			widgets.Row{
				Spacing: 8,
				Children: []gutter.Widget{
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Home", OnPressed: func() { s.miniRouter.Push("/") }},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Specs", OnPressed: func() { s.miniRouter.Push("/specs") }},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Help", OnPressed: func() { s.miniRouter.Push("/help") }},
					widgets.Button{Variant: widgets.ButtonGhost, Label: "Back", OnPressed: func() { s.miniRouter.Pop() }},
				},
			},
			widgets.Card{Variant: widgets.CardFeature, Child: widgets.RouterView{Router: s.miniRouter}},
		},
	})
}

func miniRoutePane(title, body string) gutter.Widget {
	return widgets.Column{
		Spacing: 6,
		Children: []gutter.Widget{
			widgets.Heading{Level: widgets.H5, Text: title},
			widgets.Body{Text: body, Small: true},
		},
	}
}

func authSection(s *showcaseState) gutter.Widget {
	children := []gutter.Widget{
		widgets.Body{Text: "Imported from community/login_with_google — real Google Identity Services. Set googleClientID and sign in to see the parsed JWT.", Small: true},
		loginwithgoogle.Button{
			ClientID: googleClientID,
			Text:     "continue_with",
			OnCredential: func(c loginwithgoogle.Credential) {
				s.SetState(func() {
					s.googleUser = &c
					s.googleErr = ""
				})
			},
			OnError: func(err string) {
				s.SetState(func() { s.googleErr = err })
			},
		},
		googleResultView(s),
		widgets.Row{Spacing: 12, Children: []gutter.Widget{
			widgets.Button{Variant: widgets.ButtonPrimary, Label: "Open popup", OnPressed: func() { s.popupOpen.Set(true) }},
			widgets.Button{Variant: widgets.ButtonSecondary, Label: "Open drawer", OnPressed: func() { s.drawerOpen.Set(true) }},
			widgets.Button{Variant: widgets.ButtonGhost, Label: "Open bottom sheet", OnPressed: func() { s.sheetOpen.Set(true) }},
		}},
	}
	return sectionFrame("community/login_with_google", widgets.Column{
		Spacing:  12,
		Children: children,
	})
}

func googleResultView(s *showcaseState) gutter.Widget {
	if s.googleErr != "" {
		return widgets.Card{
			Variant: widgets.CardFeature,
			Child: widgets.Body{
				Text:  "Sign-in error: " + s.googleErr,
				Small: true,
			},
		}
	}
	if s.googleUser == nil {
		return widgets.Caption{Text: "Not signed in."}
	}
	u := s.googleUser
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Row{
			CrossAxisAlign: widgets.CrossAxisCenter,
			Spacing:        16,
			Children: []gutter.Widget{
				widgets.Image{
					Src:     u.Picture,
					Width:   "56px",
					Height:  "56px",
					Rounded: "50%",
					Fit:     widgets.ImageFitCover,
				},
				widgets.Column{
					Spacing: 4,
					Children: []gutter.Widget{
						widgets.Heading{Level: widgets.H6, Text: u.Name},
						widgets.Body{Text: u.Email, Small: true},
						widgets.Caption{Text: "sub: " + u.Sub},
					},
				},
				widgets.SizedBox{Width: "24px"},
				widgets.Button{
					Variant: widgets.ButtonGhost,
					Label:   "Sign out",
					OnPressed: func() {
						s.SetState(func() {
							s.googleUser = nil
							s.googleErr = ""
						})
					},
				},
			},
		},
	}
}

func footerSection() gutter.Widget {
	return widgets.Surface{
		Variant: widgets.SurfaceAlt,
		Padding: "32px",
		Child: widgets.Row{
			MainAxisAlign: widgets.MainAxisSpaceBetween,
			Children: []gutter.Widget{
				widgets.Caption{Text: "© Gutter showcase"},
				widgets.Link{Text: "github", OnPressed: func() {}},
			},
		},
	}
}

// ============================ helpers ============================

func pickTheme() *themes.Theme {
	switch themeName {
	case "meta":
		return themes.Meta
	case "neutral":
		return themes.Neutral
	default:
		return themes.Apple
	}
}

// silence unused-import warning if no math reference makes it to compile.
var _ = math.Pi

func main() {
	gutter.RunApp(Showcase{})
}
