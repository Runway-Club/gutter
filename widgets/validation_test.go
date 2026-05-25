package widgets

import "testing"

func TestValidators(t *testing.T) {
	cases := []struct {
		name string
		v    Validator
		in   string
		want string
	}{
		{"required empty", Required("req"), "", "req"},
		{"required spaces", Required("req"), "   ", "req"},
		{"required ok", Required("req"), "x", ""},
		{"minlen short", MinLength(3, "min"), "ab", "min"},
		{"minlen ok", MinLength(3, "min"), "abc", ""},
		{"maxlen long", MaxLength(2, "max"), "abc", "max"},
		{"maxlen ok", MaxLength(2, "max"), "ab", ""},
		{"email bad", Email("em"), "nope", "em"},
		{"email ok", Email("em"), "a@b.co", ""},
		{"email empty allowed", Email("em"), "", ""},
	}
	for _, c := range cases {
		if got := c.v(c.in); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestCombineReturnsFirstError(t *testing.T) {
	v := Combine(Required("required"), MinLength(3, "too short"))
	if got := v(""); got != "required" {
		t.Errorf("empty => %q, want required", got)
	}
	if got := v("ab"); got != "too short" {
		t.Errorf("short => %q, want too short", got)
	}
	if got := v("abcd"); got != "" {
		t.Errorf("valid => %q, want empty", got)
	}
}
