// Compute kernels exposed to JS for the Go-WASM vs JS benchmark.
//
// These are the SAME algorithms, written by hand, as in kernels.js — no stdlib
// sort, no native crypto, no library. Each kernel generates its own input from
// an identical 32-bit LCG and returns a checksum so the runner can verify Go,
// TinyGo, and JS all computed the exact same thing.
//
// This represents the Go code a Gutter app runs: Gutter is only the UI layer;
// heavy work like this is plain Go compiled to WASM.
package main

import "syscall/js"

// 32-bit LCG, wraps like JS (Math.imul(...) + 12345) >>> 0.
func lcg(state *uint32) uint32 {
	*state = *state*1103515245 + 12345
	return *state
}

func mandelbrot(size int) float64 {
	const maxIter = 256
	var sum float64
	for py := 0; py < size; py++ {
		y0 := (float64(py)/float64(size))*3.0 - 1.5
		for px := 0; px < size; px++ {
			x0 := (float64(px)/float64(size))*3.0 - 2.0
			x, y := 0.0, 0.0
			iter := 0
			for x*x+y*y <= 4.0 && iter < maxIter {
				xt := x*x - y*y + x0
				y = 2*x*y + y0
				x = xt
				iter++
			}
			sum += float64(iter)
		}
	}
	return sum
}

func quicksort(a []uint32) {
	if len(a) < 2 {
		return
	}
	stack := [][2]int{{0, len(a) - 1}}
	for len(stack) > 0 {
		r := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		lo, hi := r[0], r[1]
		if lo >= hi {
			continue
		}
		pivot := a[(lo+hi)/2]
		i, j := lo, hi
		for i <= j {
			for a[i] < pivot {
				i++
			}
			for a[j] > pivot {
				j--
			}
			if i <= j {
				a[i], a[j] = a[j], a[i]
				i++
				j--
			}
		}
		if lo < j {
			stack = append(stack, [2]int{lo, j})
		}
		if i < hi {
			stack = append(stack, [2]int{i, hi})
		}
	}
}

func sortKernel(n int) float64 {
	a := make([]uint32, n)
	var st uint32 = 12345
	for i := range a {
		a[i] = lcg(&st)
	}
	quicksort(a)
	step := n / 1000
	if step < 1 {
		step = 1
	}
	var sum uint64
	for i := 0; i < n; i += step {
		sum += uint64(a[i])
	}
	return float64(sum)
}

func matmul(n int) float64 {
	a := make([]float64, n*n)
	b := make([]float64, n*n)
	c := make([]float64, n*n)
	var st uint32 = 999
	for i := range a {
		a[i] = float64(lcg(&st)) / 2147483648.0
	}
	for i := range b {
		b[i] = float64(lcg(&st)) / 2147483648.0
	}
	for i := 0; i < n; i++ {
		for k := 0; k < n; k++ {
			aik := a[i*n+k]
			for j := 0; j < n; j++ {
				c[i*n+j] += aik * b[k*n+j]
			}
		}
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += c[i*n+i]
	}
	return sum
}

func main() {
	ns := js.Global().Get("Object").New()
	ns.Set("mandelbrot", js.FuncOf(func(_ js.Value, args []js.Value) any { return mandelbrot(args[0].Int()) }))
	ns.Set("sort", js.FuncOf(func(_ js.Value, args []js.Value) any { return sortKernel(args[0].Int()) }))
	ns.Set("matmul", js.FuncOf(func(_ js.Value, args []js.Value) any { return matmul(args[0].Int()) }))
	js.Global().Set("goKernels", ns)
	js.Global().Get("document").Get("documentElement").Call("setAttribute", "data-go-ready", "1")
	select {}
}
