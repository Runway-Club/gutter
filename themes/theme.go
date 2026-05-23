// Package themes defines the data types that describe a visual theme — colors,
// typography, shape (border radius scale), spacing, and a catalog of
// pre-composed component styles — and ships the Apple and Meta design systems
// as ready-made presets. Themes are pure data; this package imports nothing
// from gutter or widgets and knows how to render nothing on its own.
//
// Pass a theme to gutter.RunApp via gutter.WithTheme(themes.Apple); widgets
// in github.com/Runway-Club/gutter/widgets/themed read the active theme from
// BuildContext to drive their styling.
package themes

// Theme is the top-level container. A theme is a static value; share the same
// pointer across the whole app.
type Theme struct {
	Name       string
	Colors     Colors
	Typography Typography
	Rounded    Rounded
	Spacing    Spacing
	Components Components
}

// Colors holds the semantic palette. Specific theme presets fill these from
// their respective design systems (Apple's parchment + ink, Meta's canvas +
// ink-deep, etc.); themed widgets reach for these by role rather than by hex.
type Colors struct {
	// Primary is the marketing-surface "click me" color. For Apple this is
	// Action Blue; for Meta this is true black (pill primary).
	Primary   string
	OnPrimary string

	// Accent is the secondary brand color. Meta uses this for the commerce-flow
	// cobalt; Apple maps it to its dark-surface link blue.
	Accent   string
	OnAccent string

	// Surfaces.
	Canvas      string // page background
	CanvasAlt   string // alternate light surface (Apple parchment, Meta soft cloud)
	SurfaceSoft string // tertiary soft surface
	SurfaceDark string // dark tile / promo strip
	OnDark      string // text on dark surfaces

	// Text on light.
	Ink       string
	InkMuted  string
	InkSubtle string

	// Hairlines / borders.
	Hairline     string
	HairlineSoft string

	// Semantic.
	Success  string
	Warning  string
	Critical string
}

// TextSpec is the CSS-ready representation of one typographic role. Empty
// fields are omitted from the rendered style.
type TextSpec struct {
	FontFamily    string
	FontSize      string
	FontWeight    string
	LineHeight    string
	LetterSpacing string
}

// Typography is the type ladder. Roles cover the common needs of marketing
// and product surfaces; both Apple and Meta map roughly to this set.
type Typography struct {
	HeroDisplay   TextSpec
	DisplayLarge  TextSpec
	DisplayMedium TextSpec
	HeadingLarge  TextSpec
	HeadingMedium TextSpec
	HeadingSmall  TextSpec
	Lead          TextSpec
	BodyStrong    TextSpec
	Body          TextSpec
	Caption       TextSpec
	CaptionStrong TextSpec
	Button        TextSpec
	Link          TextSpec
	FinePrint     TextSpec
}

// Rounded is the border-radius scale. Values are CSS strings so themes can
// express either px or % (e.g. circles).
type Rounded struct {
	None    string
	Small   string
	Medium  string
	Large   string
	XLarge  string
	XXLarge string
	Pill    string
	Circle  string
}

// Spacing is the gap/padding scale. CSS strings.
type Spacing struct {
	XXS     string
	XS      string
	SM      string
	MD      string
	LG      string
	XL      string
	XXL     string
	XXXL    string
	Section string
	Hero    string
}

// Components is the catalog of pre-composed styles for high-level widgets.
// Themed widgets read these directly so apps never write raw CSS.
type Components struct {
	ButtonPrimary   ButtonStyle
	ButtonSecondary ButtonStyle
	ButtonGhost     ButtonStyle
	ButtonAccent    ButtonStyle // Meta's cobalt buy-CTA; Apple repeats Primary
	ButtonOnDark    ButtonStyle // primary used on dark surface

	CardFeature CardStyle
	CardPromo   CardStyle
	CardPlain   CardStyle

	SurfaceCanvas SurfaceStyle
	SurfaceAlt    SurfaceStyle
	SurfaceDark   SurfaceStyle

	Input InputStyle

	BadgeNeutral  BadgeStyle
	BadgeSuccess  BadgeStyle
	BadgeWarning  BadgeStyle
	BadgeCritical BadgeStyle

	NavBar NavBarStyle
}

// NavBarStyle describes the app's top navigation strip — Scaffold's AppBar
// reads this. Apple's "global-nav" is a 44px black band; Meta's is a 64px
// white bar with a hairline separator below. Apps don't usually override it.
type NavBarStyle struct {
	Background        string
	Foreground        string
	Height            string
	Padding           string // shorthand, e.g. "0 24px"
	BorderBottomColor string
	BorderBottomWidth string
	Typography        TextSpec
}

// ButtonStyle describes a single button variant: rest colors, type, shape,
// padding, and an optional press background.
type ButtonStyle struct {
	Background       string
	Foreground       string
	BackgroundActive string
	BorderColor      string
	BorderWidth      string
	Rounded          string
	PaddingY         string
	PaddingX         string
	Typography       TextSpec
}

// CardStyle describes a component-level box: background, border, rounding,
// padding, and the text color to use for content inside.
type CardStyle struct {
	Background  string
	Foreground  string
	BorderColor string
	BorderWidth string
	Rounded     string
	Padding     string
}

// SurfaceStyle describes a layout-level region (tile, banner, hero band).
// Different from CardStyle in that it usually has no border and uses the
// section-level padding scale.
type SurfaceStyle struct {
	Background string
	Foreground string
	Padding    string
	Rounded    string
}

// InputStyle describes a text-style form control across its rest, focus, and
// error states.
type InputStyle struct {
	Background       string
	Foreground       string
	BorderColor      string
	BorderColorFocus string
	BorderColorError string
	Rounded          string
	Padding          string
	Height           string
	Typography       TextSpec
}

// BadgeStyle describes a tiny pill-shaped chip used for status indicators.
type BadgeStyle struct {
	Background string
	Foreground string
	Rounded    string
	Padding    string
	Typography TextSpec
}
