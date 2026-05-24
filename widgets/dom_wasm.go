//go:build js && wasm

package widgets

import "syscall/js"

// setBoolProp imperatively sets a boolean DOM property (e.g. "checked",
// "indeterminate"). The framework's applyAttrs only writes string attributes,
// which for inputs only sets defaultChecked / defaultValue. Setting the
// property directly is what keeps a "controlled" input in sync with its
// declarative Value/Checked field after the user has interacted with it.
func setBoolProp(node any, name string, value bool) {
	if n, ok := node.(js.Value); ok {
		n.Set(name, value)
	}
}

// setStringProp imperatively sets a string DOM property (e.g. "value" on
// inputs and selects).
func setStringProp(node any, name string, value string) {
	if n, ok := node.(js.Value); ok {
		n.Set(name, value)
	}
}

// setStringPropIfDifferent is the caret-preserving variant for text-style
// inputs and textareas. Writing `value` on every reconcile moves the caret
// to the end of the string, which makes typing in the middle of a word
// jump the cursor on every keystroke. The check makes the call a no-op
// when the DOM already holds the same string — exactly the common case
// for a controlled input echoing what the user typed.
func setStringPropIfDifferent(node any, name string, value string) {
	if n, ok := node.(js.Value); ok {
		current := n.Get(name)
		if !current.IsUndefined() && !current.IsNull() && current.String() == value {
			return
		}
		n.Set(name, value)
	}
}
