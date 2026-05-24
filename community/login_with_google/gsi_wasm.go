//go:build js && wasm

package loginwithgoogle

import "syscall/js"

const gsiScriptURL = "https://accounts.google.com/gsi/client"

// renderGoogleSignIn loads the GSI script (once per page), initializes
// Google Identity Services with the widget's ClientID, and renders the
// official "Continue with Google" button into node. Returns an idempotent
// cleanup function that releases the Go-side callback when the widget
// unmounts.
//
// Flow:
//
//  1. Insert <script src="accounts.google.com/gsi/client" async defer>
//     into <head> if not already there.
//  2. After the script's "load" event, poll briefly for
//     window.google.accounts.id (registration is normally synchronous
//     with script execution; the poll is defensive).
//  3. google.accounts.id.initialize({client_id, callback}) — callback is
//     a js.FuncOf bridge into Go.
//  4. google.accounts.id.renderButton(node, opts) to paint the button.
//  5. On user click, GSI invokes the callback with {credential: <JWT>};
//     we parse and fire onCredential.
func renderGoogleSignIn(
	node any,
	b Button,
	onCredential func(Credential),
	onError func(string),
) func() {
	n, ok := node.(js.Value)
	if !ok {
		if onError != nil {
			onError("no DOM node available for the Google button")
		}
		return func() {}
	}

	released := false
	var goCB js.Func
	goCB = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		resp := args[0]
		if resp.IsUndefined() || resp.IsNull() {
			return nil
		}
		credVal := resp.Get("credential")
		if credVal.IsUndefined() || credVal.IsNull() {
			if onError != nil {
				onError("Google response missing credential field")
			}
			return nil
		}
		parsed, err := parseJWT(credVal.String())
		if err != nil {
			if onError != nil {
				onError(err.Error())
			}
			return nil
		}
		onCredential(parsed)
		return nil
	})

	ensureGSIScript(func() {
		google := js.Global().Get("google")
		if google.IsUndefined() || google.IsNull() {
			if onError != nil {
				onError("window.google is not defined after script load")
			}
			return
		}
		gid := google.Get("accounts").Get("id")
		gid.Call("initialize", map[string]any{
			"client_id": b.ClientID,
			"callback":  goCB,
		})
		gid.Call("renderButton", n, buttonOptions(b))
	})

	return func() {
		if released {
			return
		}
		released = true
		goCB.Release()
	}
}

func buttonOptions(b Button) map[string]any {
	opts := map[string]any{
		"theme": fallback(b.Theme, "outline"),
		"size":  fallback(b.Size, "large"),
		"text":  fallback(b.Text, "continue_with"),
	}
	if b.Shape != "" {
		opts["shape"] = b.Shape
	}
	if b.Width > 0 {
		opts["width"] = b.Width
	}
	return opts
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// ensureGSIScript inserts the GSI <script> tag once per page and invokes
// ready as soon as window.google.accounts.id is defined.
func ensureGSIScript(ready func()) {
	doc := js.Global().Get("document")
	existing := doc.Call("querySelector", "script[data-gutter-gsi]")
	if !existing.IsNull() && !existing.IsUndefined() {
		pollForGSI(ready, 0)
		return
	}
	script := doc.Call("createElement", "script")
	script.Set("src", gsiScriptURL)
	script.Set("async", true)
	script.Set("defer", true)
	script.Call("setAttribute", "data-gutter-gsi", "")
	var onLoad js.Func
	onLoad = js.FuncOf(func(this js.Value, args []js.Value) any {
		onLoad.Release()
		pollForGSI(ready, 0)
		return nil
	})
	script.Call("addEventListener", "load", onLoad)
	doc.Get("head").Call("appendChild", script)
}

// pollForGSI polls (every 50 ms, capped at ~5 s) until
// window.google.accounts.id resolves. The script's `load` event normally
// fires after the library has registered itself, but ad-blockers and slow
// networks make a brief poll safer than assuming.
func pollForGSI(ready func(), attempt int) {
	google := js.Global().Get("google")
	if !google.IsUndefined() && !google.IsNull() {
		accounts := google.Get("accounts")
		if !accounts.IsUndefined() && !accounts.IsNull() {
			id := accounts.Get("id")
			if !id.IsUndefined() && !id.IsNull() {
				ready()
				return
			}
		}
	}
	if attempt > 100 {
		return // give up after ~5 s
	}
	var tick js.Func
	tick = js.FuncOf(func(this js.Value, args []js.Value) any {
		tick.Release()
		pollForGSI(ready, attempt+1)
		return nil
	})
	js.Global().Call("setTimeout", tick, 50)
}
