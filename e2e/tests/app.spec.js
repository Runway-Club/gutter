// End-to-end tests against the gutter testapp running as real WASM in a real
// browser. These cover the engine-critical flows a product built on Gutter
// depends on: first render, batched SetState, controlled inputs, keyed
// reordering, and conditional mount/unmount.
const { test, expect } = require('@playwright/test');

// waitForApp navigates to the app and waits for the WASM runtime to mount.
async function waitForApp(page) {
  await page.goto('/');
  await page.waitForFunction(
    () => document.getElementById('app') && document.getElementById('app').children.length > 0,
    null,
    { timeout: 20_000 },
  );
}

test('renders without panicking', async ({ page }) => {
  const errors = [];
  page.on('pageerror', (e) => errors.push(String(e)));
  await waitForApp(page);

  // NB: gutter's Heading renders a styled <span>, not a semantic <h1>-<h6>,
  // so we match by text rather than by ARIA role.
  await expect(page.getByText('Gutter E2E')).toBeVisible();
  expect(errors, `page errors: ${errors.join('\n')}`).toEqual([]);
});

test('counter increments on click', async ({ page }) => {
  await waitForApp(page);
  const count = page.getByTestId('count');
  await expect(count).toContainText('count: 0');
  await page.getByRole('button', { name: 'increment' }).click();
  await expect(count).toContainText('count: 1');
  await page.getByRole('button', { name: 'increment' }).click();
  await expect(count).toContainText('count: 2');
});

test('burst applies all five SetStates (batched but not lost)', async ({ page }) => {
  await waitForApp(page);
  const count = page.getByTestId('count');
  await expect(count).toContainText('count: 0');
  // One click fires five SetState calls in a single turn. Batching coalesces
  // the rebuilds, but every mutation must still apply: count must reach 5.
  await page.getByRole('button', { name: 'burst' }).click();
  await expect(count).toContainText('count: 5');
});

test('controlled input echoes its value', async ({ page }) => {
  await waitForApp(page);
  const input = page.getByPlaceholder('echo');
  await input.click();
  await input.type('hello world');
  await expect(page.getByTestId('echo')).toContainText('hello world');
  await expect(input).toHaveValue('hello world');
});

test('controlled input preserves caret on mid-string insert', async ({ page }) => {
  await waitForApp(page);
  const input = page.getByPlaceholder('echo');
  await input.click();
  await input.type('abcdef');
  // Park the caret at index 3 and type three chars one at a time. Each
  // keystroke triggers a (batched) rebuild that re-syncs value; if that reset
  // the caret to the end, Y and Z would append rather than stay contiguous.
  await input.evaluate((el) => el.setSelectionRange(3, 3));
  await page.keyboard.type('XYZ');
  await expect(input).toHaveValue('abcXYZdef');
  const caret = await input.evaluate((el) => el.selectionStart);
  expect(caret).toBe(6);
});

test('keyed reorder preserves each input value', async ({ page }) => {
  await waitForApp(page);
  // Type distinct values into A and B's inputs.
  await page.getByPlaceholder('input-A').fill('alpha');
  await page.getByPlaceholder('input-B').fill('bravo');

  // Reverse the list: A,B,C -> C,B,A. Because the rows are keyed, the existing
  // DOM nodes move rather than being recreated, so the typed values ride along.
  await page.getByRole('button', { name: 'reverse' }).click();

  await expect(page.getByPlaceholder('input-A')).toHaveValue('alpha');
  await expect(page.getByPlaceholder('input-B')).toHaveValue('bravo');

  // Confirm the visual order actually changed (C is now first).
  const labels = await page.getByTestId('keyed-list').locator('input').evaluateAll(
    (els) => els.map((e) => e.placeholder),
  );
  expect(labels).toEqual(['input-C', 'input-B', 'input-A']);
});

test('dialog mounts and unmounts on toggle', async ({ page }) => {
  await waitForApp(page);
  await expect(page.getByTestId('dialog')).toHaveCount(0);
  await page.getByRole('button', { name: 'open dialog' }).click();
  await expect(page.getByTestId('dialog')).toBeVisible();
  await expect(page.getByTestId('dialog')).toContainText('Dialog is open');
  await page.getByRole('button', { name: 'close dialog' }).click();
  await expect(page.getByTestId('dialog')).toHaveCount(0);
});
