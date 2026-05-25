// JS side of the compute benchmark — byte-for-byte the same algorithms as
// main.go. No Array.prototype.sort, no Web Crypto, no library: we compare the
// JS engine against Go-WASM on identical hand-written work. Typed arrays are
// used so V8 can optimize numeric loops (the fair, idiomatic JS baseline).

// 32-bit LCG matching Go's uint32 wrap.
function lcg(s) {
  s.v = (Math.imul(s.v, 1103515245) + 12345) >>> 0;
  return s.v;
}

const jsKernels = {
  mandelbrot(size) {
    const maxIter = 256;
    let sum = 0;
    for (let py = 0; py < size; py++) {
      const y0 = (py / size) * 3.0 - 1.5;
      for (let px = 0; px < size; px++) {
        const x0 = (px / size) * 3.0 - 2.0;
        let x = 0, y = 0, iter = 0;
        while (x * x + y * y <= 4.0 && iter < maxIter) {
          const xt = x * x - y * y + x0;
          y = 2 * x * y + y0;
          x = xt;
          iter++;
        }
        sum += iter;
      }
    }
    return sum;
  },

  sort(n) {
    const a = new Uint32Array(n);
    const s = { v: 12345 };
    for (let i = 0; i < n; i++) a[i] = lcg(s);
    // iterative quicksort, identical partition scheme to Go
    const stack = [0, n - 1];
    while (stack.length) {
      const hi = stack.pop();
      const lo = stack.pop();
      if (lo >= hi) continue;
      const pivot = a[(lo + hi) >> 1];
      let i = lo, j = hi;
      while (i <= j) {
        while (a[i] < pivot) i++;
        while (a[j] > pivot) j--;
        if (i <= j) {
          const t = a[i];
          a[i] = a[j];
          a[j] = t;
          i++;
          j--;
        }
      }
      if (lo < j) stack.push(lo, j);
      if (i < hi) stack.push(i, hi);
    }
    let step = (n / 1000) | 0;
    if (step < 1) step = 1;
    let sum = 0;
    for (let i = 0; i < n; i += step) sum += a[i];
    return sum;
  },

  matmul(n) {
    const a = new Float64Array(n * n);
    const b = new Float64Array(n * n);
    const c = new Float64Array(n * n);
    const s = { v: 999 };
    for (let i = 0; i < n * n; i++) a[i] = lcg(s) / 2147483648.0;
    for (let i = 0; i < n * n; i++) b[i] = lcg(s) / 2147483648.0;
    for (let i = 0; i < n; i++) {
      for (let k = 0; k < n; k++) {
        const aik = a[i * n + k];
        for (let j = 0; j < n; j++) {
          c[i * n + j] += aik * b[k * n + j];
        }
      }
    }
    let sum = 0;
    for (let i = 0; i < n; i++) sum += c[i * n + i];
    return sum;
  },
};

window.jsKernels = jsKernels;
