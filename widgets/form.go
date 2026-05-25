package widgets

import (
	"maps"

	"github.com/Runway-Club/gutter"
)

// FormField describes one input in a Form: its key, label, type, initial value,
// and the validators run on submit.
type FormField struct {
	Name        string
	Label       string
	Type        InputType
	Initial     string
	Placeholder string
	Validators  []Validator
}

// Form is a controlled, validated form. It owns the field values and errors,
// renders a labeled Input per field with an inline error message, and a submit
// button. On submit it runs every field's validators; OnSubmit fires with a
// snapshot of the values only when all fields pass.
//
//	widgets.Form{
//	    Fields: []widgets.FormField{
//	        {Name: "email", Label: "Email", Type: widgets.InputEmail,
//	         Validators: []widgets.Validator{
//	             widgets.Required("Email is required"),
//	             widgets.Email("Enter a valid email"),
//	         }},
//	    },
//	    Submit:   "Sign up",
//	    OnSubmit: func(v map[string]string) { /* call rpc.Call, etc. */ },
//	}
type Form struct {
	Fields   []FormField
	Submit   string // submit button label; defaults to "Submit"
	OnSubmit func(values map[string]string)
}

func (f Form) CreateState() gutter.State { return &formState{form: f} }

type formState struct {
	gutter.StateObject
	form   Form
	values map[string]string
	errors map[string]string
}

func (s *formState) InitState() {
	s.values = make(map[string]string, len(s.form.Fields))
	s.errors = map[string]string{}
	for _, fld := range s.form.Fields {
		s.values[fld.Name] = fld.Initial
	}
}

func (s *formState) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	rows := make([]gutter.Widget, 0, len(s.form.Fields)+1)
	for _, fld := range s.form.Fields {
		name := fld.Name
		children := make([]gutter.Widget, 0, 3)
		if fld.Label != "" {
			children = append(children, Body{Text: fld.Label, Bold: true})
		}
		children = append(children, Input{
			Type:        fld.Type,
			Value:       s.values[name],
			Placeholder: fld.Placeholder,
			Error:       s.errors[name] != "",
			OnChanged:   func(v string) { s.SetState(func() { s.values[name] = v }) },
		})
		if msg := s.errors[name]; msg != "" {
			children = append(children, Caption{Text: msg, Color: t.Components.Input.BorderColorError})
		}
		rows = append(rows, Column{Spacing: 4, Children: children})
	}
	rows = append(rows, Button{
		Variant:   ButtonPrimary,
		Label:     fallback(s.form.Submit, "Submit"),
		OnPressed: s.submit,
	})
	return Column{Spacing: 12, Children: rows}
}

func (s *formState) submit() {
	errs := validateFields(s.form.Fields, s.values)
	s.SetState(func() { s.errors = errs })
	if len(errs) == 0 && s.form.OnSubmit != nil {
		s.form.OnSubmit(maps.Clone(s.values))
	}
}

// validateFields runs each field's validators against the current values and
// returns field name -> first error message, for failing fields only. Pure, so
// it is unit-testable without a runtime.
func validateFields(fields []FormField, values map[string]string) map[string]string {
	errs := map[string]string{}
	for _, fld := range fields {
		v := values[fld.Name]
		for _, validate := range fld.Validators {
			if msg := validate(v); msg != "" {
				errs[fld.Name] = msg
				break
			}
		}
	}
	return errs
}
