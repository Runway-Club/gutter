package widgets

import "github.com/Runway-Club/gutter"

// AppBar is the top navigation strip used inside a Scaffold. Layout:
// [Leading] [Title (left-aligned)] ... spacer ... [Actions (right-aligned)].
// Styling — background, height, padding, typography, bottom border — comes
// from the active theme's NavBar style; apps don't need to set CSS.
//
// Pass TitleWidget to use any widget as the title instead of plain text.
type AppBar struct {
	Title       string
	TitleWidget gutter.Widget
	Leading     gutter.Widget
	Actions     []gutter.Widget
}

func (a AppBar) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := t.Components.NavBar

	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"height":           style.Height,
		"min-height":       style.Height,
		"padding":          style.Padding,
		"display":          "flex",
		"flex-direction":   "row",
		"align-items":      "center",
		"gap":              "16px",
		"width":            "100%",
		"box-sizing":       "border-box",
		"flex-shrink":      "0",
	}
	if style.BorderBottomColor != "" && style.BorderBottomWidth != "" {
		css["border-bottom"] = style.BorderBottomWidth + " solid " + style.BorderBottomColor
	}

	var children []gutter.Widget
	if a.Leading != nil {
		children = append(children, a.Leading)
	}
	title := a.TitleWidget
	if title == nil && a.Title != "" {
		titleStyle := map[string]string{"color": style.Foreground}
		applySpec(titleStyle, style.Typography)
		title = Styled{Tag: "span", Text: a.Title, Style: titleStyle}
	}
	if title != nil {
		children = append(children, title)
	}
	// Push actions to the right with a flex spacer.
	if len(a.Actions) > 0 {
		children = append(children, Styled{Style: map[string]string{"flex": "1"}})
		children = append(children, a.Actions...)
	}

	return Styled{Tag: "header", Style: css, Children: children}
}
