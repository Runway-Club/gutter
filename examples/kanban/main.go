// Kanban: a drag-and-drop board built on widgets.Draggable + DropTarget +
// DragOverlay. Three columns ("To do", "In progress", "Done"); cards move
// between them via pointer drag.
//
//	cd examples/kanban
//	go run ../../cmd/gutter run
//	# open http://localhost:8080
package main

import (
	"fmt"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
	"github.com/Runway-Club/gutter/widgets"
)

// Task is the payload moved between columns.
type Task struct {
	ID    int
	Title string
}

// columnID is just a string in this example; declaring it as a named type
// would be cleaner in real code.
const (
	colTodo  = "todo"
	colDoing = "doing"
	colDone  = "done"
)

// App is a StatefulWidget that owns all tasks and the shared drag controller.
type App struct{}

func (App) CreateState() gutter.State { return &appState{} }

type appState struct {
	gutter.StateObject
	tasks  []taskRow
	ctrl   *widgets.Controller[Task]
	nextID int
}

// taskRow couples a Task with its current column. Keeping a single flat
// list (rather than one slice per column) means a drop is just a column
// change on one row — no slice splicing, no reorder bugs.
type taskRow struct {
	Task   Task
	Column string
}

func (s *appState) InitState() {
	s.ctrl = widgets.NewController[Task]()
	seed := []struct {
		title string
		col   string
	}{
		{"Sketch the layout", colTodo},
		{"Pick a theme", colTodo},
		{"Write the docs", colTodo},
		{"Implement Draggable", colDoing},
		{"Implement DropTarget", colDoing},
		{"Wire pointer events", colDone},
		{"Ship hit-testing", colDone},
	}
	for _, t := range seed {
		s.nextID++
		s.tasks = append(s.tasks, taskRow{Task: Task{ID: s.nextID, Title: t.title}, Column: t.col})
	}
}

// tasksFor returns the tasks currently in col, in stable order (slice order).
func (s *appState) tasksFor(col string) []Task {
	out := make([]Task, 0, len(s.tasks))
	for _, r := range s.tasks {
		if r.Column == col {
			out = append(out, r.Task)
		}
	}
	return out
}

// move sets the column of task `id` to `toCol`. Re-positions the row to
// the end of its new column so dropped tasks land at the bottom.
func (s *appState) move(id int, toCol string) {
	s.SetState(func() {
		var moved taskRow
		out := make([]taskRow, 0, len(s.tasks))
		for _, r := range s.tasks {
			if r.Task.ID == id {
				moved = taskRow{Task: r.Task, Column: toCol}
				continue
			}
			out = append(out, r)
		}
		s.tasks = append(out, moved)
	})
}

func (s *appState) Build(ctx *gutter.BuildContext) gutter.Widget {
	return widgets.Scaffold{
		Title:        "Gutter Kanban",
		Theme:        themes.Apple,
		StickyAppBar: true,
		AppBar: widgets.AppBar{
			Title: "Kanban",
			Actions: []gutter.Widget{
				widgets.Caption{Text: fmt.Sprintf("%d tasks", len(s.tasks))},
			},
		},
		Body: widgets.Surface{
			Variant: widgets.SurfaceAlt,
			Padding: "32px",
			Child: widgets.Column{
				Children: []gutter.Widget{
					widgets.Row{
						Spacing:        16,
						CrossAxisAlign: widgets.CrossAxisStart,
						Children: []gutter.Widget{
							column{Ctrl: s.ctrl, Label: "To do", ID: colTodo, Tasks: s.tasksFor(colTodo), OnDrop: s.move},
							column{Ctrl: s.ctrl, Label: "In progress", ID: colDoing, Tasks: s.tasksFor(colDoing), OnDrop: s.move},
							column{Ctrl: s.ctrl, Label: "Done", ID: colDone, Tasks: s.tasksFor(colDone), OnDrop: s.move},
						},
					},
					// Mount one DragOverlay at the root — it renders the
					// ghost while a drag is active. Required for visual
					// feedback during drag.
					widgets.DragOverlay[Task]{Controller: s.ctrl},
				},
			},
		},
	}
}

// column is a StatefulWidget so each column can own its hover-highlight
// state without polluting the app's top-level State.
type column struct {
	Ctrl   *widgets.Controller[Task]
	Label  string
	ID     string
	Tasks  []Task
	OnDrop func(taskID int, toCol string)
}

func (c column) CreateState() gutter.State { return &columnState{} }

type columnState struct {
	gutter.StateObject
	over bool
}

func (s *columnState) widget() column { return s.Widget().(column) }

func (s *columnState) Build(ctx *gutter.BuildContext) gutter.Widget {
	w := s.widget()
	border := "2px dashed #d2d2d7"
	bg := "#ffffff"
	if s.over {
		border = "2px solid " + ctx.Theme.Colors.Primary
		bg = "#f0f7ff"
	}

	cards := make([]gutter.Widget, 0, len(w.Tasks)+1)
	cards = append(cards,
		widgets.Row{
			MainAxisAlign:  widgets.MainAxisSpaceBetween,
			CrossAxisAlign: widgets.CrossAxisCenter,
			Children: []gutter.Widget{
				widgets.Heading{Level: widgets.H5, Text: w.Label},
				widgets.Badge{Variant: widgets.BadgeNeutral, Text: fmt.Sprintf("%d", len(w.Tasks))},
			},
		},
	)
	for _, t := range w.Tasks {
		t := t
		cards = append(cards, widgets.Draggable[Task]{
			Controller: w.Ctrl,
			Data:       t,
			// Use the same widget as ghost for the lift-out preview.
			Ghost: cardView(t),
			Child: cardView(t),
		})
	}

	return widgets.DropTarget[Task]{
		Controller: w.Ctrl,
		OnDrop:     func(t Task) { w.OnDrop(t.ID, w.ID) },
		OnHoverChange: func(over bool) {
			s.SetState(func() { s.over = over })
		},
		// Accepts everything by default — leaving Accepts nil.
		Child: widgets.Container{
			Width:        "280px",
			Padding:      widgets.EdgeInsetsAll(16),
			Color:        bg,
			BorderRadius: "12px",
			Border:       border,
			Child: widgets.Column{
				Spacing:  12,
				Children: cards,
			},
		},
	}
}

func cardView(t Task) gutter.Widget {
	return widgets.Card{
		Variant: widgets.CardFeature,
		Child: widgets.Column{
			Spacing: 4,
			Children: []gutter.Widget{
				widgets.Body{Text: t.Title, Bold: true},
				widgets.Caption{Text: fmt.Sprintf("#%d", t.ID)},
			},
		},
	}
}

func main() { gutter.RunApp(App{}) }
