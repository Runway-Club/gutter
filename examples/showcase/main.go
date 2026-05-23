// Showcase: the exact same widget tree rendered under whichever theme is
// chosen at build time. Defaults to Apple. To render with Meta:
//
//	cd examples/showcase
//	GOOS=js GOARCH=wasm go build -ldflags "-X 'main.themeName=meta'" -o app.wasm .
package main

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// themeName is set via -ldflags. Defaults to apple.
var themeName = "apple"

type Showcase struct{}

func (Showcase) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Surface{
		Variant: widgets.SurfaceAlt,
		Padding: "0",
		Child: widgets.Column{
			Spacing: 0,
			Children: []gutter.Widget{
				// Hero band — dark surface with on-dark text + a primary CTA.
				widgets.Surface{
					Variant: widgets.SurfaceDark,
					Child: widgets.Column{
						CrossAxisAlign: widgets.CrossAxisCenter,
						Spacing:        24,
						Children: []gutter.Widget{
							widgets.Heading{Level: widgets.H1, Text: "One catalog. Two design systems.", Color: ctx.Theme.Colors.OnDark},
							widgets.Body{Text: "Compose your UI in Go. Pick a theme. Ship.", Color: ctx.Theme.Colors.OnDark},
							widgets.Row{
								Spacing: 12,
								Children: []gutter.Widget{
									widgets.Button{Variant: widgets.ButtonOnDark, Label: "Get started"},
									widgets.Button{Variant: widgets.ButtonGhost, Label: "Documentation"},
								},
							},
						},
					},
				},
				// Feature card row — three cards in the theme's CardFeature style.
				widgets.Surface{
					Variant: widgets.SurfaceCanvas,
					Child: widgets.Row{
						Spacing: 24,
						Children: []gutter.Widget{
							featureCard("Type", "A complete typographic ladder, from hero display to legal fine print, baked into the theme."),
							featureCard("Color", "Brand primary, ink, hairlines, semantic tones — every role mapped, no hex literals in app code."),
							featureCard("Shape", "Border-radius, spacing, button geometry — pick a variant, the theme picks the values."),
						},
					},
				},
				// Promo strip — dark commerce surface, accent CTA.
				widgets.Surface{
					Variant: widgets.SurfaceAlt,
					Child: widgets.Card{
						Variant: widgets.CardPromo,
						Child: widgets.Column{
							CrossAxisAlign: widgets.CrossAxisCenter,
							Spacing:        16,
							Children: []gutter.Widget{
								widgets.Heading{Level: widgets.H3, Text: "Ready to ship?", Color: ctx.Theme.Colors.OnDark},
								widgets.Body{Text: "Pre-order now and pay later.", Color: ctx.Theme.Colors.OnDark},
								widgets.Button{Variant: widgets.ButtonAccent, Label: "Add to cart"},
							},
						},
					},
				},
				// Status row — badges across all four semantic colors.
				widgets.Surface{
					Variant: widgets.SurfaceCanvas,
					Child: widgets.Row{
						Spacing:       12,
						MainAxisAlign: widgets.MainAxisCenter,
						Children: []gutter.Widget{
							widgets.Badge{Variant: widgets.BadgeNeutral, Text: "In review"},
							widgets.Badge{Variant: widgets.BadgeSuccess, Text: "In stock"},
							widgets.Badge{Variant: widgets.BadgeWarning, Text: "Selling fast"},
							widgets.Badge{Variant: widgets.BadgeCritical, Text: "Out of stock"},
						},
					},
				},
			},
		},
	}
}

func featureCard(title, body string) gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Column{
			Spacing: 8,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H5, Text: title},
				widgets.Body{Text: body, Small: true},
			},
		},
	}
}

func main() {
	theme := themes.Apple
	if themeName == "meta" {
		theme = themes.Meta
	}
	gutter.RunApp(Showcase{}, gutter.WithTheme(theme))
}
