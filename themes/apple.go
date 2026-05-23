package themes

// Apple is the design system extracted from theme_specs/APPLE_DESIGN.md.
// Photography-first, single Action Blue accent, pill buttons and rectangular
// full-bleed tiles, the signature "Apple tight" negative letter-spacing at
// display sizes.
//
// Note on font: the original spec uses SF Pro Display / SF Pro Text. For
// portability across non-Apple platforms, every TextSpec leads with Lexend
// (the gutter default font) and falls back to SF Pro and system-ui. Load
// Lexend via Google Fonts to get the intended modern look; on macOS Safari
// without Lexend, the stack falls through to SF Pro automatically.
var Apple = &Theme{
	Name: "Apple",
	Colors: Colors{
		Primary:   "#0066cc",
		OnPrimary: "#ffffff",
		Accent:    "#2997ff",
		OnAccent:  "#ffffff",

		Canvas:      "#ffffff",
		CanvasAlt:   "#f5f5f7",
		SurfaceSoft: "#fafafc",
		SurfaceDark: "#272729",
		OnDark:      "#ffffff",

		Ink:       "#1d1d1f",
		InkMuted:  "#333333",
		InkSubtle: "#7a7a7a",

		Hairline:     "#e0e0e0",
		HairlineSoft: "#f0f0f0",

		Success:  "#34c759",
		Warning:  "#ff9f0a",
		Critical: "#ff3b30",
	},
	Typography: Typography{
		HeroDisplay:   appleDisplay("56px", "600", "1.07", "-0.28px"),
		DisplayLarge:  appleDisplay("40px", "600", "1.1", "0"),
		DisplayMedium: appleText("34px", "600", "1.47", "-0.374px"),
		HeadingLarge:  appleDisplay("28px", "400", "1.14", "0.196px"),
		HeadingMedium: appleDisplay("21px", "600", "1.19", "0.231px"),
		HeadingSmall:  appleText("17px", "600", "1.24", "-0.374px"),
		Lead:          appleText("24px", "300", "1.5", "0"),
		BodyStrong:    appleText("17px", "600", "1.24", "-0.374px"),
		Body:          appleText("17px", "400", "1.47", "-0.374px"),
		Caption:       appleText("14px", "400", "1.43", "-0.224px"),
		CaptionStrong: appleText("14px", "600", "1.29", "-0.224px"),
		Button:        appleText("17px", "400", "1.0", "0"),
		Link:          appleText("17px", "400", "1.47", "-0.374px"),
		FinePrint:     appleText("12px", "400", "1.0", "-0.12px"),
	},
	Rounded: Rounded{
		None: "0px", Small: "8px", Medium: "11px", Large: "18px",
		XLarge: "18px", XXLarge: "18px", Pill: "9999px", Circle: "9999px",
	},
	Spacing: Spacing{
		XXS: "4px", XS: "8px", SM: "12px", MD: "17px", LG: "24px",
		XL: "32px", XXL: "48px", XXXL: "64px", Section: "80px", Hero: "80px",
	},
	Components: Components{
		ButtonPrimary: ButtonStyle{
			Background: "#0066cc", Foreground: "#ffffff", BackgroundActive: "#0071e3",
			Rounded: "9999px", PaddingY: "11px", PaddingX: "22px",
			Typography: appleText("17px", "400", "1.0", ""),
		},
		ButtonSecondary: ButtonStyle{
			Background: "transparent", Foreground: "#0066cc",
			BorderColor: "#0066cc", BorderWidth: "1px",
			Rounded: "9999px", PaddingY: "11px", PaddingX: "22px",
			Typography: appleText("17px", "400", "1.0", ""),
		},
		ButtonGhost: ButtonStyle{
			Background: "#fafafc", Foreground: "#333333",
			BorderColor: "#f0f0f0", BorderWidth: "1px",
			Rounded: "11px", PaddingY: "8px", PaddingX: "14px",
			Typography: appleText("14px", "400", "1.29", ""),
		},
		ButtonAccent: ButtonStyle{
			Background: "#0066cc", Foreground: "#ffffff",
			Rounded: "9999px", PaddingY: "11px", PaddingX: "22px",
			Typography: appleText("17px", "400", "1.0", ""),
		},
		ButtonOnDark: ButtonStyle{
			Background: "#1d1d1f", Foreground: "#ffffff",
			Rounded: "8px", PaddingY: "8px", PaddingX: "15px",
			Typography: appleText("14px", "400", "1.29", "-0.224px"),
		},
		CardFeature: CardStyle{
			Background: "#ffffff", Foreground: "#1d1d1f",
			BorderColor: "#e0e0e0", BorderWidth: "1px",
			Rounded: "18px", Padding: "24px",
		},
		CardPromo: CardStyle{
			Background: "#272729", Foreground: "#ffffff",
			Rounded: "0px", Padding: "80px",
		},
		CardPlain: CardStyle{
			Background: "#ffffff", Foreground: "#1d1d1f",
			Rounded: "18px", Padding: "24px",
		},
		SurfaceCanvas: SurfaceStyle{Background: "#ffffff", Foreground: "#1d1d1f", Padding: "80px"},
		SurfaceAlt:    SurfaceStyle{Background: "#f5f5f7", Foreground: "#1d1d1f", Padding: "80px"},
		SurfaceDark:   SurfaceStyle{Background: "#272729", Foreground: "#ffffff", Padding: "80px"},
		Input: InputStyle{
			Background: "#ffffff", Foreground: "#1d1d1f",
			BorderColor: "#e0e0e0", BorderColorFocus: "#0071e3", BorderColorError: "#ff3b30",
			Rounded: "9999px", Padding: "12px 20px", Height: "44px",
			Typography: appleText("17px", "400", "1.47", "-0.374px"),
		},
		BadgeNeutral:  BadgeStyle{Background: "#fafafc", Foreground: "#1d1d1f", Rounded: "9999px", Padding: "4px 10px", Typography: appleText("12px", "400", "1.0", "")},
		BadgeSuccess:  BadgeStyle{Background: "#34c759", Foreground: "#ffffff", Rounded: "9999px", Padding: "4px 10px", Typography: appleText("12px", "600", "1.0", "")},
		BadgeWarning:  BadgeStyle{Background: "#ff9f0a", Foreground: "#ffffff", Rounded: "9999px", Padding: "4px 10px", Typography: appleText("12px", "600", "1.0", "")},
		BadgeCritical: BadgeStyle{Background: "#ff3b30", Foreground: "#ffffff", Rounded: "9999px", Padding: "4px 10px", Typography: appleText("12px", "600", "1.0", "")},
		// global-nav: ultra-thin pure-black bar pinned to the top, white nav
		// links in 12px / 400 / -0.12px tracking.
		NavBar: NavBarStyle{
			Background: "#000000",
			Foreground: "#ffffff",
			Height:     "44px",
			Padding:    "0 22px",
			Typography: appleText("12px", "400", "1.0", "-0.12px"),
		},
	},
}

const appleDisplayStack = "Lexend, 'SF Pro Display', system-ui, -apple-system, sans-serif"
const appleTextStack = "Lexend, 'SF Pro Text', system-ui, -apple-system, sans-serif"

func appleDisplay(size, weight, lh, ls string) TextSpec {
	return TextSpec{FontFamily: appleDisplayStack, FontSize: size, FontWeight: weight, LineHeight: lh, LetterSpacing: ls}
}

func appleText(size, weight, lh, ls string) TextSpec {
	return TextSpec{FontFamily: appleTextStack, FontSize: size, FontWeight: weight, LineHeight: lh, LetterSpacing: ls}
}
