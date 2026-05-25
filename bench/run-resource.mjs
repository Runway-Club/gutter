// Memory + CPU companion to run.mjs.
//
// For each framework × tier it measures, via the Chrome DevTools Protocol:
//
//   Memory (after a forced GC, page idle):
//     - jsHeapMB : CDP Performance.getMetrics JSHeapUsedSize (JS heap only).
//     - wasmMB   : the Go runtime's WebAssembly linear memory (captured by
//                  monkeypatching WebAssembly.instantiate* before the app runs).
//                  This is where Gutter's Go heap lives — it is NOT in the JS
//                  heap — so it is the dominant part of Gutter's footprint and
//                  is invisible to JSHeapUsedSize. React has none (0).
//     - totalMB  : jsHeapMB + wasmMB — the comparable footprint number.
//     - nodes    : live DOM node count.
//
//   CPU (main-thread busy time, from CDP Performance.getMetrics cumulative
//   counters; durations are seconds, converted to ms):
//     - renderTask/renderScript/renderLayout : work accumulated from navigation
//       until render-complete = the cost of starting up + building the tree.
//     - per-update script/layout µs : delta across a burst of BURST clicks
//       (one item updated per click), divided by the click count.
import { chromium } from 'playwright';
import { startServer } from './server.mjs';
import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const TIERS = [10, 100, 1000, 10000];
const FRAMEWORKS = [
  { id: 'gutter', label: 'Gutter (Go WASM)', mount: '/gutter' },
  { id: 'gutter-tinygo', label: 'Gutter (TinyGo WASM)', mount: '/gutter-tinygo' },
  { id: 'react', label: 'React (Vite prod)', mount: '/react' },
];
const BURST = 120; // clicks per update-CPU sample
const SETTLE_MS = 500;

const INIT_SCRIPT = () => {
  // Capture the WASM linear-memory object as it is instantiated, so we can read
  // its byteLength later. Works for both Go (exports.mem) and TinyGo
  // (exports.memory) by scanning for any WebAssembly.Memory export.
  window.__wasmMem = null;
  const grab = (result) => {
    try {
      const inst = result && (result.instance || result);
      if (inst && inst.exports) {
        for (const k in inst.exports) {
          if (inst.exports[k] instanceof WebAssembly.Memory) {
            window.__wasmMem = inst.exports[k];
            break;
          }
        }
      }
    } catch (e) {}
    return result;
  };
  const oInst = WebAssembly.instantiate.bind(WebAssembly);
  WebAssembly.instantiate = (...a) => oInst(...a).then(grab);
  if (WebAssembly.instantiateStreaming) {
    const oStream = WebAssembly.instantiateStreaming.bind(WebAssembly);
    WebAssembly.instantiateStreaming = (...a) => oStream(...a).then(grab);
  }

  const n = parseInt(new URLSearchParams(location.search).get('n') || '100', 10);
  window.__benchN = n;
  (function poll() {
    if (document.querySelectorAll('#grid [data-bench-item]').length >= n) {
      document.documentElement.setAttribute('data-ready', '1');
      return;
    }
    requestAnimationFrame(poll);
  })();
  // Click `k` items (cycling), waiting a frame after each so the framework
  // flushes its update + layout before the next — realistic per-update cost.
  window.__burst = async function (k) {
    const els = [...document.querySelectorAll('#grid [data-bench-item]')];
    if (!els.length) return;
    for (let i = 0; i < k; i++) {
      els[i % els.length].click();
      await new Promise((r) => requestAnimationFrame(r));
    }
  };
};

async function getMetrics(cdp) {
  const { metrics } = await cdp.send('Performance.getMetrics');
  return Object.fromEntries(metrics.map((m) => [m.name, m.value]));
}

async function forceGC(cdp) {
  try {
    await cdp.send('HeapProfiler.collectGarbage');
  } catch {}
}

function median(xs) {
  const s = [...xs].sort((a, b) => a - b);
  const m = Math.floor(s.length / 2);
  return s.length % 2 ? s[m] : (s[m - 1] + s[m]) / 2;
}

async function measureCell(browser, baseUrl, fw, n) {
  const timeout = n >= 10000 ? 120000 : 30000;
  const url = `${baseUrl}${fw.mount}/?n=${n}`;

  const ctx = await browser.newContext();
  await ctx.addInitScript(INIT_SCRIPT);
  const page = await ctx.newPage();
  const cdp = await ctx.newCDPSession(page);
  await cdp.send('Performance.enable');

  await page.goto(url, { waitUntil: 'load', timeout });
  await page.waitForFunction(
    () => document.documentElement.getAttribute('data-ready') === '1',
    null,
    { timeout },
  );
  await page.waitForTimeout(SETTLE_MS);

  // --- CPU of initial load + render (cumulative since navigation) ---
  const afterRender = await getMetrics(cdp);

  // --- memory at idle, after GC ---
  await forceGC(cdp);
  await page.waitForTimeout(150);
  const memMetrics = await getMetrics(cdp);
  const wasmBytes = await page.evaluate(() =>
    window.__wasmMem ? window.__wasmMem.buffer.byteLength : 0,
  );

  // --- per-update CPU: snapshot, burst, snapshot ---
  const before = await getMetrics(cdp);
  await page.evaluate((k) => window.__burst(k), BURST);
  await page.waitForTimeout(100);
  const after = await getMetrics(cdp);

  await cdp.detach();
  await ctx.close();

  const dScript = (after.ScriptDuration - before.ScriptDuration) / BURST;
  const dLayout =
    (after.LayoutDuration - before.LayoutDuration + after.RecalcStyleDuration - before.RecalcStyleDuration) /
    BURST;

  const jsHeapMB = (memMetrics.JSHeapUsedSize || 0) / 1048576;
  const wasmMB = wasmBytes / 1048576;
  return {
    jsHeapMB,
    wasmMB,
    totalMB: jsHeapMB + wasmMB,
    nodes: memMetrics.Nodes || 0,
    renderTaskMs: (afterRender.TaskDuration || 0) * 1000,
    renderScriptMs: (afterRender.ScriptDuration || 0) * 1000,
    renderLayoutMs: ((afterRender.LayoutDuration || 0) + (afterRender.RecalcStyleDuration || 0)) * 1000,
    updScriptUs: dScript * 1e6,
    updLayoutUs: dLayout * 1e6,
  };
}

function num(x, d = 1) {
  return Number.isFinite(x) && x >= 0 ? x.toFixed(d) : 'n/a';
}

async function main() {
  const wanted = process.argv[2] ? process.argv[2].split(',').map(Number) : TIERS;
  const { server, port } = await startServer(0);
  const baseUrl = `http://localhost:${port}`;
  const browser = await chromium.launch();

  const results = {};
  for (const fw of FRAMEWORKS) {
    results[fw.id] = {};
    for (const n of wanted) {
      process.stdout.write(`measuring ${fw.label}  n=${n} ... `);
      try {
        // two passes; report the lower-noise median-ish (min) for memory, mean for CPU
        const a = await measureCell(browser, baseUrl, fw, n);
        const b = await measureCell(browser, baseUrl, fw, n);
        const cell = {
          totalMB: median([a.totalMB, b.totalMB]),
          jsHeapMB: median([a.jsHeapMB, b.jsHeapMB]),
          wasmMB: median([a.wasmMB, b.wasmMB]),
          nodes: a.nodes,
          renderTaskMs: median([a.renderTaskMs, b.renderTaskMs]),
          renderScriptMs: median([a.renderScriptMs, b.renderScriptMs]),
          renderLayoutMs: median([a.renderLayoutMs, b.renderLayoutMs]),
          updScriptUs: median([a.updScriptUs, b.updScriptUs]),
          updLayoutUs: median([a.updLayoutUs, b.updLayoutUs]),
        };
        results[fw.id][n] = cell;
        console.log(
          `total ${num(cell.totalMB)}MB (wasm ${num(cell.wasmMB)} + js ${num(
            cell.jsHeapMB,
          )})  render-cpu ${num(cell.renderTaskMs)}ms  upd ${num(cell.updScriptUs, 0)}+${num(
            cell.updLayoutUs,
            0,
          )}µs`,
        );
      } catch (err) {
        console.log(`FAILED: ${err.message}`);
        results[fw.id][n] = { error: err.message };
      }
    }
  }

  await browser.close();
  server.close();

  const out = { generatedAt: new Date().toISOString(), tiers: wanted, results };
  await fs.writeFile(path.join(__dirname, 'resources.json'), JSON.stringify(out, null, 2));
  await fs.writeFile(path.join(__dirname, 'RESOURCES.md'), renderMarkdown(out));
  console.log('\nwrote resources.json and RESOURCES.md');
}

function renderMarkdown(out) {
  const present = FRAMEWORKS.filter((f) => out.results[f.id]);
  let md = `# Gutter vs React — memory & CPU\n\n`;
  md += `_Generated ${out.generatedAt}. Tiers = item count. Median of 2 passes._\n\n`;

  const tables = [
    ['totalMB', 'Total memory (MB) — JS heap + WASM linear memory, after GC', 1],
    ['wasmMB', 'WASM linear memory (MB) — Go heap; React has none', 1],
    ['jsHeapMB', 'JS heap used (MB)', 2],
    ['nodes', 'Live DOM nodes', 0],
    ['renderTaskMs', 'Cold render CPU — main-thread busy (ms)', 1],
    ['renderScriptMs', 'Cold render CPU — script/WASM exec (ms)', 1],
    ['renderLayoutMs', 'Cold render CPU — layout + style (ms)', 1],
    ['updScriptUs', 'Per-update script CPU (µs)', 0],
    ['updLayoutUs', 'Per-update layout+style CPU (µs)', 0],
  ];
  for (const [key, title, dp] of tables) {
    md += `## ${title}\n\n`;
    md += `| Items | ${present.map((f) => f.label).join(' | ')} |\n`;
    md += `|--:|${present.map(() => '--:').join('|')}|\n`;
    for (const n of out.tiers) {
      const cells = present.map((f) => {
        const c = out.results[f.id][n];
        if (!c || c.error) return 'n/a';
        return num(c[key], dp);
      });
      md += `| ${n} | ${cells.join(' | ')} |\n`;
    }
    md += `\n`;
  }
  return md;
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
