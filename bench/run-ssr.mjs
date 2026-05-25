// SSR PoC measurement: First/Largest Contentful Paint for the same dashboard
// rendered client-side (CSR, boots WASM into an empty #app) vs server-side
// (SSR, HTML pre-rendered by gutter.RenderToHTML). Cold load, fresh cache.
import { chromium } from 'playwright';
import { startServer } from './server.mjs';

const VARIANTS = [
  { id: 'CSR (client WASM render)', mount: '/ssr-csr' },
  { id: 'SSR (RenderToHTML)', mount: '/ssr-ssr' },
];
const PASSES = 5;

const INIT = () => {
  window.__lcp = 0;
  try {
    new PerformanceObserver((l) => {
      const es = l.getEntries();
      if (es.length) window.__lcp = es[es.length - 1].startTime;
    }).observe({ type: 'largest-contentful-paint', buffered: true });
  } catch (e) {}
};

function median(xs) {
  const s = [...xs].sort((a, b) => a - b);
  return s[s.length >> 1];
}

async function measure(browser, baseUrl, mount) {
  const fcps = [];
  const lcps = [];
  for (let p = 0; p < PASSES; p++) {
    const ctx = await browser.newContext();
    await ctx.addInitScript(INIT);
    const page = await ctx.newPage();
    await page.goto(`${baseUrl}${mount}/`, { waitUntil: 'load', timeout: 30000 });
    await page.waitForTimeout(600); // let WASM (CSR) finish + LCP settle
    const m = await page.evaluate(() => {
      const paint = performance.getEntriesByType('paint');
      const fcp = (paint.find((p) => p.name === 'first-contentful-paint') || {}).startTime || 0;
      return { fcp, lcp: window.__lcp || 0 };
    });
    fcps.push(m.fcp);
    lcps.push(m.lcp);
    await ctx.close();
  }
  return { fcp: median(fcps), lcp: median(lcps) };
}

async function main() {
  const { server, port } = await startServer(0);
  const baseUrl = `http://localhost:${port}`;
  const browser = await chromium.launch();

  console.log(`SSR PoC — median of ${PASSES} cold loads\n`);
  const rows = [];
  for (const v of VARIANTS) {
    const r = await measure(browser, baseUrl, v.mount);
    rows.push({ ...v, ...r });
    console.log(`${v.id.padEnd(28)}  FCP ${r.fcp.toFixed(1)}ms   LCP ${r.lcp.toFixed(1)}ms`);
  }

  if (rows.length === 2) {
    const [csr, ssr] = rows;
    const fcpGain = (csr.fcp / ssr.fcp).toFixed(1);
    console.log(`\nSSR paints first content ${fcpGain}× sooner (FCP ${csr.fcp.toFixed(0)}ms → ${ssr.fcp.toFixed(0)}ms).`);
  }

  // Hydration check: the SSR page ships static "Likes: 0"; after WASM boots and
  // hydrates the existing DOM, clicking must increment it.
  await verifyHydration(browser, baseUrl);

  await browser.close();
  server.close();
}

async function verifyHydration(browser, baseUrl) {
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto(`${baseUrl}/ssr-ssr/`, { waitUntil: 'load', timeout: 30000 });
  const btn = page.getByText(/Likes:/);
  const before = await btn.textContent();
  // Wait until hydration has wired the click (poll: click then check change).
  let after = before;
  for (let i = 0; i < 40 && after === before; i++) {
    await btn.click();
    await page.waitForTimeout(50);
    after = await btn.textContent();
  }
  await ctx.close();
  if (after !== before) {
    console.log(`Hydration OK: counter interactive after boot ("${before.trim()}" → "${after.trim()}").`);
  } else {
    console.log(`Hydration FAILED: counter still "${before.trim()}" after clicks.`);
    process.exitCode = 1;
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
