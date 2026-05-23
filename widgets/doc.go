// Package widgets is the standard widget catalog. Widgets fall into two
// loose groups:
//
//   - Themed widgets read styling from the active theme on BuildContext
//     (set by gutter.RunApp; defaults to themes.Apple). App code does not
//     write CSS — pick a variant and the theme picks the values:
//     Heading, Body, Caption, Link, Button, Card, Surface, Input, Badge.
//
//   - Layout and primitive widgets carry no theme dependency. Use them as
//     building blocks: Text, Container, Column, Row, Center, Padding,
//     SizedBox, Styled (the arbitrary-CSS escape hatch), WithKey.
//
// Typical use:
//
//	gutter.RunApp(MyApp{})  // defaults to themes.Apple + Lexend
//	// or
//	gutter.RunApp(MyApp{}, gutter.WithTheme(themes.Meta))
//
//	// Inside Build — no CSS in sight:
//	widgets.Card{
//	    Variant: widgets.CardFeature,
//	    Child: widgets.Column{
//	        Spacing: 16,
//	        Children: []gutter.Widget{
//	            widgets.Heading{Level: widgets.H1, Text: "Welcome"},
//	            widgets.Body{Text: "Pick a theme and ship."},
//	            widgets.Button{
//	                Variant:   widgets.ButtonPrimary,
//	                Label:     "Get started",
//	                OnPressed: onStart,
//	            },
//	        },
//	    },
//	}
package widgets
