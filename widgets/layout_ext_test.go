package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter"
)

func TestExpandedDefaultFlex(t *testing.T) {
	h := hostOf(t, Expanded{Child: Text{Data: "x"}})
	wantStyle(t, h, "flex", "1 1 0%")
	wantStyle(t, h, "min-width", "0")
	wantStyle(t, h, "min-height", "0")
	wantChildren(t, h, 1)
}

func TestExpandedExplicitFlex(t *testing.T) {
	h := hostOf(t, Expanded{Flex: 3, Child: Text{Data: "x"}})
	wantStyle(t, h, "flex", "3 1 0%")
}

func TestFlexible(t *testing.T) {
	h := hostOf(t, Flexible{Flex: 2, Child: Text{Data: "x"}})
	wantStyle(t, h, "flex", "2 1 auto")
}

func TestSpacer(t *testing.T) {
	h := hostOf(t, Spacer{})
	wantStyle(t, h, "flex", "1 1 0%")
	wantChildren(t, h, 0)
}

func TestStack(t *testing.T) {
	h := hostOf(t, Stack{
		Width:    "100px",
		Height:   "50px",
		Children: []gutter.Widget{Text{Data: "base"}, Positioned{Child: Text{Data: "over"}}},
	})
	wantStyle(t, h, "position", "relative")
	wantStyle(t, h, "width", "100px")
	wantStyle(t, h, "height", "50px")
	wantChildren(t, h, 2)
}

func TestPositioned(t *testing.T) {
	h := hostOf(t, Positioned{Top: "0", Right: "8px", Child: Text{Data: "x"}})
	wantStyle(t, h, "position", "absolute")
	wantStyle(t, h, "top", "0")
	wantStyle(t, h, "right", "8px")
	wantNoStyle(t, h, "bottom")
	wantNoStyle(t, h, "left")
}

func TestPositionedFill(t *testing.T) {
	h := hostOf(t, Positioned{Fill: true, Child: Text{Data: "x"}})
	wantStyle(t, h, "position", "absolute")
	wantStyle(t, h, "inset", "0")
}

func TestGridFixedColumns(t *testing.T) {
	h := hostOf(t, Grid{Columns: 3, Gap: 16, Children: []gutter.Widget{Text{Data: "a"}}})
	wantStyle(t, h, "display", "grid")
	wantStyle(t, h, "grid-template-columns", "repeat(3, 1fr)")
	wantStyle(t, h, "gap", "16px")
}

func TestGridResponsiveMinColumnWidth(t *testing.T) {
	h := hostOf(t, Grid{MinColumnWidth: "140px"})
	wantStyle(t, h, "grid-template-columns", "repeat(auto-fill, minmax(140px, 1fr))")
}

func TestGridTemplateWinsOverColumns(t *testing.T) {
	// Template has the highest precedence.
	h := hostOf(t, Grid{Columns: 4, MinColumnWidth: "100px", Template: "1fr 2fr"})
	wantStyle(t, h, "grid-template-columns", "1fr 2fr")
}

func TestWrap(t *testing.T) {
	h := hostOf(t, Wrap{Spacing: 8, RunSpacing: 4, Children: []gutter.Widget{Text{Data: "a"}}})
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "flex-wrap", "wrap")
	wantStyle(t, h, "flex-direction", "row")
	// gap is "<row-gap=RunSpacing> <column-gap=Spacing>" for a row.
	wantStyle(t, h, "gap", "4px 8px")
}

func TestAlignPresets(t *testing.T) {
	h := hostOf(t, Align{Alignment: AlignBottomRight, Child: Text{Data: "x"}})
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "justify-content", MainAxisEnd)
	wantStyle(t, h, "align-items", CrossAxisEnd)
	wantStyle(t, h, "width", "100%")
	wantStyle(t, h, "height", "100%")
}

func TestAlignZeroValueDefaultsToCenter(t *testing.T) {
	h := hostOf(t, Align{Child: Text{Data: "x"}})
	wantStyle(t, h, "justify-content", MainAxisCenter)
	wantStyle(t, h, "align-items", CrossAxisCenter)
}

func TestAspectRatio(t *testing.T) {
	h := hostOf(t, AspectRatio{Ratio: 16.0 / 9.0, Child: Text{Data: "x"}})
	wantStyle(t, h, "aspect-ratio", "1.7777777777777777")
	wantStyle(t, h, "width", "100%")
	wantStyle(t, h, "overflow", "hidden")
}

func TestAspectRatioZeroDefaultsToSquare(t *testing.T) {
	h := hostOf(t, AspectRatio{})
	wantStyle(t, h, "aspect-ratio", "1")
}

func TestConstrainedBox(t *testing.T) {
	h := hostOf(t, ConstrainedBox{MaxWidth: "720px", MinHeight: "100px", Child: Text{Data: "x"}})
	wantStyle(t, h, "max-width", "720px")
	wantStyle(t, h, "min-height", "100px")
	wantNoStyle(t, h, "min-width")
	wantNoStyle(t, h, "max-height")
}
