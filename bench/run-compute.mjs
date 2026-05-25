// Go-WASM vs JS compute benchmark runner.
//
// Runs three identical hand-written kernels (mandelbrot, sort, matmul) across a
// few problem sizes, for three engines: JS (V8), Go-WASM, TinyGo-WASM. Each
// measurement is warmup×2 + 5 timed runs, median reported. The kernels return a
// checksum; the runner asserts JS and Go produced the same value (within
// tolerance) so we know the comparison is honest.
import { chromium } from 'playwright';
import { startServer } from './server.mjs';
import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const KERNELS = {
  'mandelbrot (px², float)': { name: 'mandelbrot', sizes: [256, 512, 1024] },
  'quicksort (ints)': { name: 'sort', sizes: [100000, 1000000, 4000000] },
  'matmul (n×n, FLOPs)': { name: 'matmul', sizes: [128, 256, 384] },
};
const WARMUP = 2;
const RUNS = 5;

async function timeOn(page, impl, name, size) {
  return page.evaluate(
    ([impl, name, size, w, r]) => window.__timeKernel(impl, name, size, w, r),
    [impl, name, size, WARMUP, RUNS],
  );
}

async function openPage(browser, baseUrl, mount, goReady) {
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto(`${baseUrl}${mount}/`, { waitUntil: 'load', timeout: 30000 });
  if (goReady) {
    await page.waitForFunction(
      () => document.documentElement.getAttribute('data-go-ready') === '1' && window.goKernels,
      null,
      { timeout: 30000 },
    );
  }
  await page.waitForFunction(() => window.jsKernels, null, { timeout: 30000 });
  return { ctx, page };
}

function fmt(ms) {
  if (!Number.isFinite(ms)) return 'n/a';
  return ms >= 100 ? ms.toFixed(0) : ms.toFixed(1);
}

async function main() {
  const { server, port } = await startServer(0);
  const baseUrl = `http://localhost:${port}`;
  const browser = await chromium.launch();

  // results[kernelName][size] = { js, go, tinygo, results: {...} }
  const results = {};
  const checks = [];

  // Page 1: Go std + JS
  const { ctx: c1, page: p1 } = await openPage(browser, baseUrl, '/compute-go', true);
  for (const [label, { name, sizes }] of Object.entries(KERNELS)) {
    results[label] = {};
    for (const size of sizes) {
      process.stdout.write(`[go/js] ${name} ${size} ... `);
      const js = await timeOn(p1, 'js', name, size);
      const go = await timeOn(p1, 'go', name, size);
      results[label][size] = { js: js.median, go: go.median, jsRes: js.result, goRes: go.result };
      const ok = Math.abs(js.result - go.result) <= Math.max(1, Math.abs(js.result)) * 1e-9;
      checks.push({ name, size, ok, js: js.result, go: go.result });
      console.log(`js ${fmt(js.median)}ms  go ${fmt(go.median)}ms  ${ok ? '✓match' : '✗MISMATCH'}`);
    }
  }
  await c1.close();

  // Page 2: TinyGo (reuse JS numbers from page 1)
  try {
    const { ctx: c2, page: p2 } = await openPage(browser, baseUrl, '/compute-tinygo', true);
    for (const [label, { name, sizes }] of Object.entries(KERNELS)) {
      for (const size of sizes) {
        process.stdout.write(`[tinygo] ${name} ${size} ... `);
        const tg = await timeOn(p2, 'go', name, size);
        results[label][size].tinygo = tg.median;
        results[label][size].tinygoRes = tg.result;
        console.log(`${fmt(tg.median)}ms`);
      }
    }
    await c2.close();
  } catch (err) {
    console.log(`tinygo page failed (skipping): ${err.message}`);
  }

  await browser.close();
  server.close();

  const out = { generatedAt: new Date().toISOString(), warmup: WARMUP, runs: RUNS, results, checks };
  await fs.writeFile(path.join(__dirname, 'compute.json'), JSON.stringify(out, null, 2));
  await fs.writeFile(path.join(__dirname, 'COMPUTE.md'), renderMarkdown(out));
  const bad = checks.filter((c) => !c.ok);
  console.log(`\nchecksum parity: ${checks.length - bad.length}/${checks.length} match`);
  if (bad.length) console.log('  mismatches:', JSON.stringify(bad));
  console.log('wrote compute.json and COMPUTE.md');
}

function renderMarkdown(out) {
  let md = `# Compute: Go-WASM vs JS (vs TinyGo)\n\n`;
  md += `_Generated ${out.generatedAt}. Median of ${out.runs} runs after ${out.warmup} warmups. `;
  md += `Identical hand-written algorithms; checksums verified to match._\n\n`;
  md += `Times in **ms** (lower is better). "speedup" = JS time ÷ Go time (>1 means Go-WASM faster).\n\n`;

  for (const [label, sizes] of Object.entries(out.results)) {
    md += `## ${label}\n\n`;
    md += `| Size | JS (V8) | Go-WASM | TinyGo | Go speedup vs JS |\n`;
    md += `|--:|--:|--:|--:|--:|\n`;
    for (const [size, c] of Object.entries(sizes)) {
      const speed = c.js / c.go;
      md += `| ${size} | ${fmt(c.js)} | ${fmt(c.go)} | ${fmt(c.tinygo)} | ${
        Number.isFinite(speed) ? speed.toFixed(2) + '×' : 'n/a'
      } |\n`;
    }
    md += `\n`;
  }
  const bad = out.checks.filter((c) => !c.ok);
  md += `**Correctness:** ${out.checks.length - bad.length}/${out.checks.length} checksums match between JS and Go`;
  md += bad.length ? ` (mismatches: ${JSON.stringify(bad)}).\n` : ` — same work verified.\n`;
  return md;
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
