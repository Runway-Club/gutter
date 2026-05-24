package widgets

import (
	"strconv"
	"sync/atomic"

	"github.com/Runway-Club/gutter"
)

// RadioOption is one entry in a RadioGroup, mirroring SelectOption.
type RadioOption[T any] struct {
	Value T
	Label string
}

// RadioGroup is a set of native <input type="radio"> buttons sharing a name
// so the browser enforces single-selection. The group is generic over the
// option value type, so apps work with typed enums or domain values rather
// than raw strings.
//
//	widgets.RadioGroup[string]{
//	    Options: []widgets.RadioOption[string]{
//	        {Value: "s",  Label: "Small"},
//	        {Value: "m",  Label: "Medium"},
//	        {Value: "l",  Label: "Large"},
//	    },
//	    Selected:  size,
//	    OnChanged: func(s string) { size = s },
//	}
//
// Direction picks the layout axis ("row" or "column"; defaults to "column").
// A stable name attribute is generated once per mounted group via InitState
// so multiple groups on the same page don't bleed into one another.
type RadioGroup[T comparable] struct {
	Options   []RadioOption[T]
	Selected  T
	OnChanged func(T)
	Disabled  bool
	Direction string
}

func (g RadioGroup[T]) CreateState() gutter.State { return &radioGroupState[T]{} }

type radioGroupState[T comparable] struct {
	gutter.StateObject
	name string
}

var radioGroupCounter uint64

func (s *radioGroupState[T]) InitState() {
	id := atomic.AddUint64(&radioGroupCounter, 1)
	s.name = "gutter-radio-" + strconv.FormatUint(id, 10)
}

func (s *radioGroupState[T]) currentWidget() RadioGroup[T] { return s.Widget().(RadioGroup[T]) }

func (s *radioGroupState[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	g := s.currentWidget()
	t := activeTheme(ctx)
	direction := g.Direction
	if direction == "" {
		direction = "column"
	}

	selectedIndex := -1
	for i, opt := range g.Options {
		if opt.Value == g.Selected {
			selectedIndex = i
			break
		}
	}

	children := make([]gutter.Widget, 0, len(g.Options))
	for i, opt := range g.Options {
		i := i
		labelStyle := map[string]string{
			"display":      "inline-flex",
			"align-items":  "center",
			"gap":          "8px",
			"cursor":       "pointer",
			"user-select":  "none",
			"color":        t.Colors.Ink,
			"accent-color": t.Colors.Primary,
		}
		if g.Disabled {
			labelStyle["opacity"] = "0.6"
			labelStyle["cursor"] = "not-allowed"
		}
		applySpec(labelStyle, t.Components.Input.Typography)

		attrs := map[string]string{
			"type":  "radio",
			"name":  s.name,
			"value": strconv.Itoa(i),
		}
		if i == selectedIndex {
			attrs["checked"] = ""
		}
		if g.Disabled {
			attrs["disabled"] = ""
		}

		checked := (i == selectedIndex)
		index := i
		input := propSyncHost{
			tag:   "input",
			attrs: attrs,
			style: map[string]string{"width": "18px", "height": "18px", "margin": "0"},
			onMount: func(node any) {
				setBoolProp(node, "checked", checked)
			},
		}
		if g.OnChanged != nil && !g.Disabled {
			oc := g.OnChanged
			input.events = map[string]func(gutter.Event){
				"change": func(gutter.Event) {
					oc(g.Options[index].Value)
				},
			}
		}

		children = append(children, Styled{
			Tag:   "label",
			Style: labelStyle,
			Children: []gutter.Widget{
				input,
				Styled{Text: opt.Label},
			},
		})
	}

	gap := "12px"
	if direction == "row" {
		gap = "16px"
	}
	return Styled{
		Style: map[string]string{
			"display":        "flex",
			"flex-direction": direction,
			"gap":            gap,
		},
		Children: children,
	}
}
