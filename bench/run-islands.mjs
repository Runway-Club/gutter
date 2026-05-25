// Smoke test for the islands example: a static HTML page that lazy-loads
// app.wasm on first island visibility, then MountInto wires two independent
// islands. Verifies both islands become interactive and stay independent.
import { chromium } from 'playwright';
import { startServer } from './server.mjs';

async function main() {
  const { server, port } = await startServer(0);
  const browser = await chromium.launch();
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto(`http://localhost:${port}/islands/`, { waitUntil: 'load', timeout: 30000 });

  // The first island is near the top → loader boots WASM → island mounts.
  const cart = page.getByText(/Add to cart/);
  await cart.waitFor({ timeout: 30000 });
  const cart0 = await cart.textContent();
  await cart.click();
  await page.waitForTimeout(80);
  const cart1 = await cart.textContent();

  // Second island (below a tall spacer) is independent.
  const likes = page.getByText(/Like/);
  const likes0 = await likes.textContent();

  await browser.close();
  server.close();

  const cartWorks = cart0 !== cart1;
  const independent = /\(0\)/.test(likes0 || '');
  console.log(`island 1 interactive: ${cartWorks ? 'OK' : 'FAIL'} ("${cart0?.trim()}" → "${cart1?.trim()}")`);
  console.log(`island 2 independent: ${independent ? 'OK' : 'FAIL'} ("${likes0?.trim()}")`);
  if (!cartWorks || !independent) process.exitCode = 1;
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
