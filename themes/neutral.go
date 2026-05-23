package themes

// Neutral is a Lexend-based fallback theme with no brand opinions: neutral
// blue accent, system grayscale ink, sensible spacing, conservative rounding.
// Apps that don't care about a specific brand can use this; the gutter
// runtime falls back to it only when used outside of a normal RunApp call
// (e.g. unit tests). RunApp itself defaults to Apple.
var Neutral = &Theme{
	Name: "Neutral",
	Colors: Colors{
		Primary:   "#1f6feb",
		OnPrimary: "#ffffff",
		Accent:    "#1f6feb",
		OnAccent:  "#ffffff",

		Canvas:      "#ffffff",
		CanvasAlt:   "#f6f8fa",
		SurfaceSoft: "#f6f8fa",
		SurfaceDark: "#0d1117",
		OnDark:      "#ffffff",

		Ink:       "#1f2328",
		InkMuted:  "#656d76",
		InkSubtle: "#8b949e",

		Hairline:     "#d0d7de",
		HairlineSoft: "#eaeef2",

		Success:  "#1a7f37",
		Warning:  "#bf8700",
		Critical: "#d1242f",
	},
	Typography: Typography{
		HeroDisplay:   lex("48px", "700", "1.1", ""),
		DisplayLarge:  lex("36px", "700", "1.15", ""),
		DisplayMedium: lex("28px", "600", "1.2", ""),
		HeadingLarge:  lex("24px", "600", "1.25", ""),
		HeadingMedium: lex("20px", "600", "1.3", ""),
		HeadingSmall:  lex("16px", "600", "1.4", ""),
		Lead:          lex("18px", "400", "1.5", ""),
		BodyStrong:    lex("16px", "600", "1.5", ""),
		Body:          lex("16px", "400", "1.5", ""),
		Caption:       lex("14px", "400", "1.4", ""),
		CaptionStrong: lex("14px", "600", "1.4", ""),
		Button:        lex("14px", "600", "1.0", ""),
		Link:          lex("16px", "500", "1.5", ""),
		FinePrint:     lex("12px", "400", "1.4", ""),
	},
	Rounded: Rounded{
		None: "0px", Small: "4px", Medium: "6px", Large: "10px",
		XLarge: "14px", XXLarge: "20px", Pill: "9999px", Circle: "9999px",
	},
	Spacing: Spacing{
		XXS: "4px", XS: "8px", SM: "12px", MD: "16px", LG: "20px",
		XL: "24px", XXL: "32px", XXXL: "40px", Section: "64px", Hero: "96px",
	},
	Components: Components{
		ButtonPrimary: ButtonStyle{
			Background: "#1f6feb", Foreground: "#ffffff",
			Rounded: "10px", PaddingY: "8px", PaddingX: "16px",
			Typography: lex("14px", "600", "1.0", ""),
		},
		ButtonSecondary: ButtonStyle{
			Background: "transparent", Foreground: "#1f6feb",
			BorderColor: "#1f6feb", BorderWidth: "1px",
			Rounded: "10px", PaddingY: "8px", PaddingX: "16px",
			Typography: lex("14px", "600", "1.0", ""),
		},
		ButtonGhost: ButtonStyle{
			Background: "transparent", Foreground: "#1f2328",
			BorderColor: "#d0d7de", BorderWidth: "1px",
			Rounded: "10px", PaddingY: "8px", PaddingX: "16px",
			Typography: lex("14px", "600", "1.0", ""),
		},
		ButtonAccent: ButtonStyle{
			Background: "#1a7f37", Foreground: "#ffffff",
			Rounded: "10px", PaddingY: "8px", PaddingX: "16px",
			Typography: lex("14px", "600", "1.0", ""),
		},
		ButtonOnDark: ButtonStyle{
			Background: "#ffffff", Foreground: "#1f2328",
			Rounded: "10px", PaddingY: "8px", PaddingX: "16px",
			Typography: lex("14px", "600", "1.0", ""),
		},
		CardFeature: CardStyle{
			Background: "#ffffff", Foreground: "#1f2328",
			BorderColor: "#d0d7de", BorderWidth: "1px",
			Rounded: "14px", Padding: "20px",
		},
		CardPromo: CardStyle{
			Background: "#0d1117", Foreground: "#ffffff",
			Rounded: "14px", Padding: "32px",
		},
		CardPlain: CardStyle{
			Background: "#ffffff", Foreground: "#1f2328",
			Rounded: "14px", Padding: "20px",
		},
		SurfaceCanvas: SurfaceStyle{Background: "#ffffff", Foreground: "#1f2328", Padding: "64px"},
		SurfaceAlt:    SurfaceStyle{Background: "#f6f8fa", Foreground: "#1f2328", Padding: "64px"},
		SurfaceDark:   SurfaceStyle{Background: "#0d1117", Foreground: "#ffffff", Padding: "64px"},
		Input: InputStyle{
			Background: "#ffffff", Foreground: "#1f2328",
			BorderColor: "#d0d7de", BorderColorFocus: "#1f6feb", BorderColorError: "#d1242f",
			Rounded: "8px", Padding: "10px 12px", Height: "40px",
			Typography: lex("16px", "400", "1.5", ""),
		},
		BadgeNeutral:  BadgeStyle{Background: "#eaeef2", Foreground: "#1f2328", Rounded: "9999px", Padding: "2px 8px", Typography: lex("12px", "600", "1.0", "")},
		BadgeSuccess:  BadgeStyle{Background: "#1a7f37", Foreground: "#ffffff", Rounded: "9999px", Padding: "2px 8px", Typography: lex("12px", "600", "1.0", "")},
		BadgeWarning:  BadgeStyle{Background: "#bf8700", Foreground: "#ffffff", Rounded: "9999px", Padding: "2px 8px", Typography: lex("12px", "600", "1.0", "")},
		BadgeCritical: BadgeStyle{Background: "#d1242f", Foreground: "#ffffff", Rounded: "9999px", Padding: "2px 8px", Typography: lex("12px", "600", "1.0", "")},
		NavBar: NavBarStyle{
			Background:        "#ffffff",
			Foreground:        "#1f2328",
			Height:            "56px",
			Padding:           "0 20px",
			BorderBottomColor: "#d0d7de",
			BorderBottomWidth: "1px",
			Typography:        lex("14px", "600", "1.0", ""),
		},
	},
}

// LexendStack is the framework-default font stack. Lexend is loaded from
// Google Fonts in the scaffolded index.html; the rest of the stack provides
// graceful fallback when Lexend isn't available.
const LexendStack = "Lexend, system-ui, -apple-system, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"

func lex(size, weight, lh, ls string) TextSpec {
	return TextSpec{FontFamily: LexendStack, FontSize: size, FontWeight: weight, LineHeight: lh, LetterSpacing: ls}
}
