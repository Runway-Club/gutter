//go:build !js || !wasm

package loginwithgoogle

// renderGoogleSignIn is a no-op on host builds. Google Identity Services
// only exists in a browser, so the button can't paint anywhere meaningful
// off the WASM target. The stub keeps user code compiling for editor
// tooling and `go vet` on the host.
func renderGoogleSignIn(
	node any,
	b Button,
	onCredential func(Credential),
	onError func(string),
) func() {
	return func() {}
}
