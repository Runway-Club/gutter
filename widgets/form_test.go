package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter/themes"
)

func TestValidateFields(t *testing.T) {
	fields := []FormField{
		{Name: "email", Validators: []Validator{Required("required"), Email("invalid")}},
		{Name: "name", Validators: []Validator{Required("required")}},
	}

	all := validateFields(fields, map[string]string{})
	if all["email"] != "required" || all["name"] != "required" {
		t.Fatalf("empty form errors = %v", all)
	}

	mixed := validateFields(fields, map[string]string{"email": "bad", "name": "Jo"})
	if mixed["email"] != "invalid" {
		t.Errorf("email error = %q, want invalid", mixed["email"])
	}
	if _, ok := mixed["name"]; ok {
		t.Errorf("name should validate, got error %q", mixed["name"])
	}

	valid := validateFields(fields, map[string]string{"email": "a@b.co", "name": "Jo"})
	if len(valid) != 0 {
		t.Errorf("valid form reported errors: %v", valid)
	}
}

func TestFormInitStateSeedsValues(t *testing.T) {
	f := Form{Fields: []FormField{{Name: "x", Label: "X", Initial: "hi"}}}
	st := f.CreateState().(*formState)
	st.InitState()
	if st.values["x"] != "hi" {
		t.Errorf("initial value = %q, want hi", st.values["x"])
	}
	// Build must not panic and returns a widget tree.
	if w := st.Build(testCtx(themes.Apple)); w == nil {
		t.Fatal("Build returned nil")
	}
}
