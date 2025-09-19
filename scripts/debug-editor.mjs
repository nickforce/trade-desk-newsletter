import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';

const OUT = (name) => path.join(process.cwd(), 'out', name);

(async () => {
  const browser = await chromium.launch({ headless: false, args: ['--no-sandbox','--disable-setuid-sandbox'] });
  const context = await browser.newContext({ storageState: path.join(process.cwd(), 'secrets', 'substack_state.json') });
  const page = await context.newPage();

  await page.goto('https://substack.com/publish', { waitUntil: 'domcontentloaded' });
  await page.screenshot({ path: OUT('publish.png'), fullPage: true });

  const frames = page.frames().map(f => ({ url: f.url(), name: f.name() }));
  fs.writeFileSync(OUT('frames.json'), JSON.stringify(frames, null, 2));

  // Try to click New post
  const np = await page.getByRole('link', { name: /new post|write/i }).first();
  if (await np.isVisible().catch(()=>false)) await np.click();
  else {
    const nb = await page.getByRole('button', { name: /new post|write/i }).first();
    if (await nb.isVisible().catch(()=>false)) await nb.click();
  }
  await page.waitForTimeout(1500);
  await page.screenshot({ path: OUT('editor.png'), fullPage: true });

  console.log('Saved out/publish.png, out/editor.png, out/frames.json');
})();
