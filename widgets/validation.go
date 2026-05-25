package widgets

import (
	"regexp"
	"strings"
)

// Validator checks a field value and returns an error message, or "" if the
// value is valid. Validators are plain functions, so they compose freely and
// can be shared with server-side checks (e.g. alongside a gutter/rpc handler).
type Validator func(value string) string

// Required fails with msg when the value is empty (after trimming whitespace).
func Required(msg string) Validator {
	return func(v string) string {
		if strings.TrimSpace(v) == "" {
			return msg
		}
		return ""
	}
}

// MinLength fails with msg when the value has fewer than n runes.
func MinLength(n int, msg string) Validator {
	return func(v string) string {
		if len([]rune(v)) < n {
			return msg
		}
		return ""
	}
}

// MaxLength fails with msg when the value has more than n runes.
func MaxLength(n int, msg string) Validator {
	return func(v string) string {
		if len([]rune(v)) > n {
			return msg
		}
		return ""
	}
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Email fails with msg when a non-empty value isn't a plausible email address.
// Empty is allowed — combine with Required to also enforce presence.
func Email(msg string) Validator {
	return func(v string) string {
		if v != "" && !emailRe.MatchString(v) {
			return msg
		}
		return ""
	}
}

// Pattern fails with msg when a non-empty value doesn't match re.
func Pattern(re *regexp.Regexp, msg string) Validator {
	return func(v string) string {
		if v != "" && !re.MatchString(v) {
			return msg
		}
		return ""
	}
}

// Combine runs validators in order and returns the first non-empty error.
func Combine(validators ...Validator) Validator {
	return func(v string) string {
		for _, validate := range validators {
			if msg := validate(v); msg != "" {
				return msg
			}
		}
		return ""
	}
}
