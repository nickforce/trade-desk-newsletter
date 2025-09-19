// scripts/open-editor.mjs
import { chromium } from 'playwright';
import path from 'path';

const ROOT = process.cwd();
const STATE_PATH =
  process.env.SUBSTACK_STATE_PATH || path.join(ROOT, 'secrets', 'substack_state.json');

// You can override the URL when running, defaults to new post editor
const targetUrl =
  process.env.SUBSTACK_EDITOR_URL ||
  'https://nickjaguarvision.substack.com/publish/post?type=newsletter';

(async () => {
  const browser = await chromium.launch({
    headless: false, // show the browser
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
  });
  const ctx = await browser.newContext({ storageState: STATE_PATH });
  const page = await ctx.newPage();

  console.log(`\n🟢 Opening ${targetUrl}`);
  await page.goto(targetUrl, { waitUntil: 'domcontentloaded' });

  console.log('\n🔎 Browser is paused.');
  console.log('   Use the Playwright Inspector crosshair to click the element you want (title, subtitle, body, or publish buttons).');
  console.log('   The Inspector will show you the exact locator string.\n');

  // Pause here so you can interact and pick locators with Inspector
  await page.pause();
})();
