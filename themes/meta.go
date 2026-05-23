package themes

// Meta is the design system extracted from theme_specs/META_DESIGN.md.
// Stark white canvas with full-bleed product photography, dual-CTA pattern
// (black pill primary on marketing + cobalt pill on commerce), 100px pill
// buttons and 32px rounded photographic cards.
//
// Note on font: the original spec uses Optimistic VF. Every TextSpec leads
// with Lexend (the gutter default font) and falls back to Optimistic VF /
// Montserrat / sans-serif. Load Lexend via Google Fonts.
var Meta = &Theme{
	Name: "Meta",
	Colors: Colors{
		Primary:   "#000000",
		OnPrimary: "#ffffff",
		Accent:    "#0064e0",
		OnAccent:  "#ffffff",

		Canvas:      "#ffffff",
		CanvasAlt:   "#f1f4f7",
		SurfaceSoft: "#f1f4f7",
		SurfaceDark: "#0a1317",
		OnDark:      "#ffffff",

		Ink:       "#1c1e21",
		InkMuted:  "#444950",
		InkSubtle: "#5d6c7b",

		Hairline:     "#ced0d4",
		HairlineSoft: "#dee3e9",

		Success:  "#31a24c",
		Warning:  "#f2a918",
		Critical: "#e41e3f",
	},
	Typography: Typography{
		HeroDisplay:   metaSpec("64px", "500", "1.16", ""),
		DisplayLarge:  metaSpec("48px", "500", "1.17", ""),
		DisplayMedium: metaSpec("36px", "500", "1.28", ""),
		HeadingLarge:  metaSpec("28px", "300", "1.21", ""),
		HeadingMedium: metaSpec("24px", "500", "1.25", ""),
		HeadingSmall:  metaSpec("18px", "700", "1.44", ""),
		Lead:          metaSpec("18px", "400", "1.44", ""),
		BodyStrong:    metaSpec("16px", "700", "1.50", "-0.16px"),
		Body:          metaSpec("16px", "400", "1.50", "-0.16px"),
		Caption:       metaSpec("12px", "400", "1.33", ""),
		CaptionStrong: metaSpec("12px", "700", "1.33", ""),
		Button:        metaSpec("14px", "700", "1.43", "-0.14px"),
		Link:          metaSpec("16px", "700", "1.50", "-0.16px"),
		FinePrint:     metaSpec("12px", "400", "1.33", ""),
	},
	Rounded: Rounded{
		None: "0px", Small: "4px", Medium: "8px", Large: "16px",
		XLarge: "24px", XXLarge: "32px", Pill: "100px", Circle: "9999px",
	},
	Spacing: Spacing{
		XXS: "4px", XS: "8px", SM: "10px", MD: "12px", LG: "20px",
		XL: "24px", XXL: "32px", XXXL: "40px", Section: "64px", Hero: "120px",
	},
	Components: Components{
		ButtonPrimary: ButtonStyle{
			Background: "#000000", Foreground: "#ffffff", BackgroundActive: "#444950",
			Rounded: "100px", PaddingY: "14px", PaddingX: "30px",
			Typography: metaSpec("14px", "700", "1.43", "-0.14px"),
		},
		ButtonSecondary: ButtonStyle{
			Background: "transparent", Foreground: "#0a1317",
			BorderColor: "#0a1317", BorderWidth: "2px",
			Rounded: "100px", PaddingY: "12px", PaddingX: "28px",
			Typography: metaSpec("14px", "700", "1.43", "-0.14px"),
		},
		ButtonGhost: ButtonStyle{
			Background: "transparent", Foreground: "#0a1317",
			BorderColor: "rgba(10, 19, 23, 0.12)", BorderWidth: "2px",
			Rounded: "100px", PaddingY: "10px", PaddingX: "22px",
			Typography: metaSpec("14px", "700", "1.43", "-0.14px"),
		},
		ButtonAccent: ButtonStyle{
			Background: "#0064e0", Foreground: "#ffffff", BackgroundActive: "#0457cb",
			Rounded: "100px", PaddingY: "14px", PaddingX: "30px",
			Typography: metaSpec("14px", "700", "1.43", "-0.14px"),
		},
		ButtonOnDark: ButtonStyle{
			Background: "#ffffff", Foreground: "#0a1317",
			Rounded: "100px", PaddingY: "14px", PaddingX: "30px",
			Typography: metaSpec("14px", "700", "1.43", "-0.14px"),
		},
		CardFeature: CardStyle{
			Background: "#ffffff", Foreground: "#1c1e21",
			BorderColor: "#dee3e9", BorderWidth: "1px",
			Rounded: "32px", Padding: "32px",
		},
		CardPromo: CardStyle{
			Background: "#0a1317", Foreground: "#ffffff",
			Rounded: "32px", Padding: "64px",
		},
		CardPlain: CardStyle{
			Background: "#ffffff", Foreground: "#1c1e21",
			BorderColor: "#dee3e9", BorderWidth: "1px",
			Rounded: "16px", Padding: "24px",
		},
		SurfaceCanvas: SurfaceStyle{Background: "#ffffff", Foreground: "#1c1e21", Padding: "64px"},
		SurfaceAlt:    SurfaceStyle{Background: "#f1f4f7", Foreground: "#1c1e21", Padding: "64px"},
		SurfaceDark:   SurfaceStyle{Background: "#0a1317", Foreground: "#ffffff", Padding: "64px"},
		Input: InputStyle{
			Background: "#ffffff", Foreground: "#1c1e21",
			BorderColor: "#ced0d4", BorderColorFocus: "#1876f2", BorderColorError: "#f0284a",
			Rounded: "8px", Padding: "12px", Height: "44px",
			Typography: metaSpec("16px", "400", "1.50", "-0.16px"),
		},
		BadgeNeutral:  BadgeStyle{Background: "#f1f4f7", Foreground: "#1c1e21", Rounded: "100px", Padding: "4px 10px", Typography: metaSpec("12px", "700", "1.33", "")},
		BadgeSuccess:  BadgeStyle{Background: "#31a24c", Foreground: "#ffffff", Rounded: "100px", Padding: "4px 10px", Typography: metaSpec("12px", "700", "1.33", "")},
		BadgeWarning:  BadgeStyle{Background: "#f2a918", Foreground: "#0a1317", Rounded: "100px", Padding: "4px 10px", Typography: metaSpec("12px", "700", "1.33", "")},
		BadgeCritical: BadgeStyle{Background: "#e41e3f", Foreground: "#ffffff", Rounded: "100px", Padding: "4px 10px", Typography: metaSpec("12px", "700", "1.33", "")},
		// Top nav: white canvas, ~64px tall, 1px hairline-soft bottom border,
		// body-sm-bold typography to match the pill-tab navigation labels.
		NavBar: NavBarStyle{
			Background:        "#ffffff",
			Foreground:        "#1c1e21",
			Height:            "64px",
			Padding:           "0 32px",
			BorderBottomColor: "#dee3e9",
			BorderBottomWidth: "1px",
			Typography:        metaSpec("14px", "700", "1.43", "-0.14px"),
		},
	},
}

const metaStack = "Lexend, 'Optimistic VF', Montserrat, Helvetica, Arial, sans-serif"

func metaSpec(size, weight, lh, ls string) TextSpec {
	return TextSpec{FontFamily: metaStack, FontSize: size, FontWeight: weight, LineHeight: lh, LetterSpacing: ls}
}
