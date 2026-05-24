//go:build js && wasm

package widgets

import "syscall/js"

// attachFileChangeListener wires a "change" listener to an <input type="file">
// DOM node. When the user picks files, each is read into memory via
// FileReader, and onSelect is invoked once all reads have completed.
//
// Returns a cleanup func that removes the listener and releases the js.Func
// allocations. The cleanup is idempotent.
func attachFileChangeListener(node any, onSelect func([]FilePick)) func() {
	n, ok := node.(js.Value)
	if !ok || onSelect == nil {
		return func() {}
	}
	released := false
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, _ []js.Value) any {
		fileList := n.Get("files")
		if fileList.IsUndefined() || fileList.IsNull() {
			return nil
		}
		count := fileList.Length()
		if count == 0 {
			return nil
		}
		results := make([]FilePick, count)
		remaining := count
		for i := 0; i < count; i++ {
			i := i
			file := fileList.Index(i)
			results[i] = FilePick{
				Name:     file.Get("name").String(),
				Size:     int64(file.Get("size").Float()),
				MimeType: file.Get("type").String(),
			}
			reader := js.Global().Get("FileReader").New()
			var loadCB js.Func
			loadCB = js.FuncOf(func(this js.Value, _ []js.Value) any {
				buf := reader.Get("result")
				arr := js.Global().Get("Uint8Array").New(buf)
				data := make([]byte, arr.Get("length").Int())
				js.CopyBytesToGo(data, arr)
				results[i].Data = data
				remaining--
				if remaining == 0 {
					onSelect(results)
				}
				loadCB.Release()
				return nil
			})
			reader.Call("addEventListener", "load", loadCB)
			reader.Call("readAsArrayBuffer", file)
		}
		return nil
	})
	n.Call("addEventListener", "change", cb)
	return func() {
		if released {
			return
		}
		released = true
		n.Call("removeEventListener", "change", cb)
		cb.Release()
	}
}
