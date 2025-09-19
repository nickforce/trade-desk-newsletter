import fs from 'fs';
import path from 'path';
import { chromium } from 'playwright';

const ROOT = process.cwd();
const STATE_PATH =
  process.env.SUBSTACK_STATE_PATH || path.join(ROOT, 'secrets', 'substack_state.json');
const MODE = process.env.SUBSTACK_MODE || 'draft'; // 'publish' | 'draft'
const TITLE_PREFIX = process.env.POST_TITLE_PREFIX || 'In Play — ';
const SUBTITLE_TEXT =
  process.env.SUBSTACK_SUBTITLE ||
  `Daily desk notes — ${new Date().toISOString().slice(0, 10)}`;
const HEADLESS =
  (process.env.SUBSTACK_HEADLESS ?? 'true').toLowerCase() !== 'false';

// ----- args & content -----
const mdPath = process.argv[2];
if (!mdPath || !fs.existsSync(mdPath)) {
  console.error('Usage: node scripts/post-substack.mjs out/daily-YYYY-MM-DD.md');
  process.exit(1);
}
const content = fs.readFileSync(mdPath, 'utf8');
const base = path.basename(mdPath).replace(/\.md$/, '');
const dateStr = base.replace(/^daily-/, '');
const title = `${TITLE_PREFIX}${dateStr}`;

(async () => {
  const browser = await chromium.launch({
    headless: HEADLESS,
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
  });
  const context = await browser.newContext({ storageState: STATE_PATH });
  const page = await context.newPage();

  // 1) Go straight to editor
  const editorUrl =
    'https://nickjaguarvision.substack.com/publish/post?type=newsletter';
  await page.goto(editorUrl, {
    waitUntil: 'domcontentloaded',
    timeout: 60000,
  });

  // 2) Fill Title
  const titleEl = page.getByTestId('post-title');
  await titleEl.waitFor({ state: 'visible', timeout: 20000 });
  await titleEl.click();
  await titleEl.fill(title);

  // 3) Fill Subtitle (optional)
  const subtitleEl = page.getByRole('textbox', { name: 'Add a subtitle…' });
  if (await subtitleEl.isVisible().catch(() => false)) {
    await subtitleEl.click();
    await subtitleEl.fill(SUBTITLE_TEXT);
  }

  // 4) Fill Body
  const bodyEl = page.getByRole('paragraph').first();
  await bodyEl.waitFor({ state: 'visible', timeout: 20000 });
  await bodyEl.click();
  await page.keyboard.type(content, { delay: 0 });

  // give autosave a few seconds
  await page.waitForTimeout(3000);

  // 5) Publish if requested
  if (MODE === 'publish') {
    const cont = page.getByRole('button', { name: /continue/i }).first();
    if (await cont.isVisible().catch(() => false)) {
      console.log('👉 Clicking Continue…');
      await cont.click();
      await page.waitForTimeout(2000);

      // Actual publish button in your account
      const pubNow = page.getByRole('button', { name: 'Send to everyone now' }).first();
      if (await pubNow.isVisible().catch(() => false)) {
        console.log('👉 Clicking Send to everyone now…');
        await pubNow.click();
        // wait extra to ensure publish completes
        await page.waitForTimeout(7000);
      } else {
        console.warn('⚠️ Send to everyone now button not found, left as draft.');
      }
    } else {
      console.warn('⚠️ "Continue" not found, left as draft.');
    }
  }

  console.log(
    `✅ ${MODE === 'publish' ? 'Published' : 'Draft created'}: ${title}`
  );
  await browser.close();
})().catch(async (err) => {
  console.error('ERROR:', err.message || err);
  process.exit(1);
});
