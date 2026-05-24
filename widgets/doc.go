// Package widgets is the standard widget catalog. Widgets fall into three
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
//   - Imperative and lifecycle widgets reach beyond the declarative DOM:
//     Canvas (custom 2D painting via a typed painter), GestureDetector
//     (pointer/keyboard event hooks for any child), and Worker (offloads
//     heavy work to a Web Worker with a builder API).
//
//   - Reactive control-flow widgets wrap async or observable inputs:
//     ObserverBuilder (subscribes to a gutter.Listenable and rebuilds on
//     change), AsyncBuilder (runs a func returning a value+error in a
//     goroutine and rebuilds with the snapshot), Router and RouterView
//     (path-based routing on top of the browser history API).
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
