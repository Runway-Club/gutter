//go:build !js || !wasm

package widgets

func setBoolProp(node any, name string, value bool)                {}
func setStringProp(node any, name string, value string)            {}
func setStringPropIfDifferent(node any, name string, value string) {}
