// Gutter vs React render/reload benchmark runner.
//
// For each framework × each tier (item count) it measures:
//   1. Cold initial render — fresh browser context (empty cache): FCP, LCP, and
//      "render-complete" (time from navigation until all N items are in #grid).
//   2. Warm reload — same context, page.reload(): cache is primed, so this
//      isolates runtime startup (WASM instantiate / JS parse+exec) from network.
//   3. Update latency — click one item's button, measure click→DOM-update via an
//      in-page MutationObserver; median of several clicks.
//   4. Bundle size — raw + gzipped bytes of each app's shipped files.
//
// All page-side instrumentation is injected identically into both apps via
// addInitScript, so the methodology can't diverge between frameworks.
import { chromium } from 'playwright';
import { startServer } from './server.mjs';
import { promises as fs } from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const TIERS = [10, 100, 1000, 10000];
const FRAMEWORKS = [
  { id: 'gutter', label: 'Gutter (Go CSR)', mount: '/gutter', distDir: 'gutter-app/dist' },
  { id: 'gutter-tinygo', label: 'Gutter (TinyGo CSR)', mount: '/gutter-tinygo', distDir: 'gutter-app/dist-tinygo' },
  { id: 'gutter-ssr', label: 'Gutter (SSR, Go)', mount: '/gutter-ssr', distDir: 'gutter-app/dist-ssr' },
  { id: 'gutter-ssr-tinygo', label: 'Gutter (SSR, TinyGo)', mount: '/gutter-ssr-tinygo', distDir: 'gutter-app/dist-ssr-tinygo' },
  { id: 'react', label: 'React (Vite prod)', mount: '/react', distDir: 'react-app/dist' },
];

// The SSR variant is pre-rendered per tier (n{N}.html); ?n= tells the wasm which
// tree to hydrate. Everyone else reads ?n= into a single page.
function urlFor(baseUrl, fw, n) {
  if (fw.id.startsWith('gutter-ssr')) return `${baseUrl}${fw.mount}/n${n}.html?n=${n}`;
  return `${baseUrl}${fw.mount}/?n=${n}`;
}
const CLICKS = 15; // update-latency samples per cell
const SETTLE_MS = 400; // allow LCP/layout to settle before reading metrics

// Injected into every page before any app code runs — identical for both apps.
const INIT_SCRIPT = () => {
  window.__lcp = 0;
  try {
    new PerformanceObserver((list) => {
      const es = list.getEntries();
      if (es.length) window.__lcp = es[es.length - 1].startTime;
    }).observe({ type: 'largest-contentful-paint', buffered: true });
  } catch (e) {}

  const n = parseInt(new URLSearchParams(location.search).get('n') || '100', 10);
  window.__benchN = n;
  window.__appReady = 0;
  (function poll() {
    if (document.querySelectorAll('#grid [data-bench-item]').length >= n) {
      window.__appReady = performance.now();
      document.documentElement.setAttribute('data-ready', '1');
      return;
    }
    requestAnimationFrame(poll);
  })();

  // click→DOM-update latency for a single item, measured entirely in-page.
  window.__measureUpdate = function (selector) {
    return new Promise((resolve) => {
      const el = document.querySelector(selector);
      if (!el) return resolve(-1);
      const t0 = performance.now();
      const obs = new MutationObserver(() => {
        obs.disconnect();
        requestAnimationFrame(() => resolve(performance.now() - t0));
      });
      obs.observe(el, { childList: true, subtree: true, characterData: true });
      el.click();
    });
  };
};

function median(xs) {
  if (!xs.length) return NaN;
  const s = [...xs].sort((a, b) => a - b);
  const m = Math.floor(s.length / 2);
  return s.length % 2 ? s[m] : (s[m - 1] + s[m]) / 2;
}

async function readMetrics(page) {
  return page.evaluate(() => {
    const paint = performance.getEntriesByType('paint');
    const fcp = (paint.find((p) => p.name === 'first-contentful-paint') || {}).startTime || 0;
    const nav = performance.getEntriesByType('navigation')[0] || {};
    return {
      fcp,
      lcp: window.__lcp || 0,
      ready: window.__appReady || 0,
      domContentLoaded: nav.domContentLoadedEventEnd || 0,
      transferBytes: (performance.getEntriesByType('resource') || []).reduce(
        (a, r) => a + (r.transferSize || 0),
        0,
      ),
    };
  });
}

async function waitReady(page, timeout) {
  await page.waitForFunction(() => document.documentElement.getAttribute('data-ready') === '1', null, {
    timeout,
  });
}

async function measureCell(browser, baseUrl, fw, n) {
  const timeout = n >= 10000 ? 120000 : 30000;
  const url = urlFor(baseUrl, fw, n);

  // --- cold load: fresh context = empty cache ---
  const ctx = await browser.newContext();
  await ctx.addInitScript(INIT_SCRIPT);
  const page = await ctx.newPage();
  await page.goto(url, { waitUntil: 'load', timeout });
  await waitReady(page, timeout);
  await page.waitForTimeout(SETTLE_MS);
  const cold = await readMetrics(page);

  // --- warm reload: cache is now primed ---
  await page.reload({ waitUntil: 'load', timeout });
  await waitReady(page, timeout);
  await page.waitForTimeout(SETTLE_MS);
  const warm = await readMetrics(page);

  // --- update latency: click one item's button repeatedly ---
  const updates = [];
  for (let k = 0; k < CLICKS; k++) {
    const dt = await page.evaluate(() => window.__measureUpdate('#grid [data-bench-item="0"]'));
    if (dt >= 0) updates.push(dt);
    await page.waitForTimeout(20);
  }

  await ctx.close();

  return {
    coldFcp: cold.fcp,
    coldLcp: cold.lcp,
    coldReady: cold.ready,
    coldTransfer: cold.transferBytes,
    warmReady: warm.ready,
    warmFcp: warm.fcp,
    updateMedian: median(updates),
  };
}

async function bundleSize(distDir) {
  const root = path.join(__dirname, distDir);
  let raw = 0;
  let gz = 0;
  const files = [];
  async function walk(dir) {
    let entries;
    try {
      entries = await fs.readdir(dir, { withFileTypes: true });
    } catch {
      return;
    }
    for (const e of entries) {
      const p = path.join(dir, e.name);
      if (e.isDirectory()) {
        await walk(p);
      } else if (!e.name.endsWith('.html')) {
        // HTML is per-page content (esp. SSR's pre-rendered pages), not the
        // shippable code bundle — exclude it so sizes compare like-for-like.
        const buf = await fs.readFile(p);
        raw += buf.length;
        gz += zlib.gzipSync(buf, { level: 9 }).length;
        files.push({ name: path.relative(root, p), raw: buf.length });
      }
    }
  }
  await walk(root);
  return { raw, gz, files };
}

function kb(bytes) {
  return (bytes / 1024).toFixed(1) + ' KB';
}
function ms(x) {
  return Number.isFinite(x) ? x.toFixed(1) : 'n/a';
}

async function main() {
  const wanted = process.argv[2] ? process.argv[2].split(',').map(Number) : TIERS;
  const { server, port } = await startServer(0);
  const baseUrl = `http://localhost:${port}`;
  const browser = await chromium.launch();

  const results = {};
  const sizes = {};

  for (const fw of FRAMEWORKS) {
    sizes[fw.id] = await bundleSize(fw.distDir);
    if (sizes[fw.id].raw === 0) {
      console.log(`! skipping ${fw.label}: no build found at ${fw.distDir}`);
      continue;
    }
    results[fw.id] = {};
    for (const n of wanted) {
      process.stdout.write(`measuring ${fw.label}  n=${n} ... `);
      try {
        const cell = await measureCell(browser, baseUrl, fw, n);
        results[fw.id][n] = cell;
        console.log(
          `cold-ready ${ms(cell.coldReady)}ms  warm-ready ${ms(cell.warmReady)}ms  upd ${ms(
            cell.updateMedian,
          )}ms`,
        );
      } catch (err) {
        console.log(`FAILED: ${err.message}`);
        results[fw.id][n] = { error: err.message };
      }
    }
  }

  await browser.close();
  server.close();

  const out = { generatedAt: new Date().toISOString(), tiers: wanted, sizes, results };
  await fs.writeFile(path.join(__dirname, 'results.json'), JSON.stringify(out, null, 2));
  await fs.writeFile(path.join(__dirname, 'RESULTS.md'), renderMarkdown(out));
  console.log('\nwrote results.json and RESULTS.md');
}

function renderMarkdown(out) {
  const fwById = Object.fromEntries(FRAMEWORKS.map((f) => [f.id, f]));
  const present = FRAMEWORKS.filter((f) => out.results[f.id]);
  let md = `# Gutter vs React — render/reload benchmark\n\n`;
  md += `_Generated ${out.generatedAt}. Tiers = item count (one stateful \`<button>\` per item)._\n\n`;

  md += `## Bundle size (what the browser downloads)\n\n`;
  md += `| Framework | Raw | Gzipped |\n|---|--:|--:|\n`;
  for (const f of present) {
    md += `| ${f.label} | ${kb(out.sizes[f.id].raw)} | ${kb(out.sizes[f.id].gz)} |\n`;
  }
  md += `\n_Bundle is fixed per framework; it does not grow with item count._\n\n`;

  const metrics = [
    ['coldReady', 'Cold render-complete (ms) — nav → all items in DOM, empty cache'],
    ['warmReady', 'Warm render-complete (ms) — reload with primed cache'],
    ['coldFcp', 'Cold First Contentful Paint (ms)'],
    ['coldLcp', 'Cold Largest Contentful Paint (ms)'],
    ['updateMedian', 'Update latency (ms) — click → DOM update, median'],
  ];
  for (const [key, title] of metrics) {
    md += `## ${title}\n\n`;
    md += `| Items | ${present.map((f) => f.label).join(' | ')} |\n`;
    md += `|--:|${present.map(() => '--:').join('|')}|\n`;
    for (const n of out.tiers) {
      const cells = present.map((f) => {
        const c = out.results[f.id][n];
        if (!c || c.error) return 'n/a';
        return ms(c[key]);
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
