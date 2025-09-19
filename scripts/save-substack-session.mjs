import { chromium } from 'playwright';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const STATE_PATH = path.resolve(__dirname, '..', 'secrets', 'substack_state.json');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  await page.goto('https://substack.com', { waitUntil: 'domcontentloaded' });

  console.log('\n🟢 A browser window is open. Please log in to Substack manually (magic link or password).');
  console.log('   Once you’re on your publication dashboard, come back here. The script will wait.\n');

  // Give yourself 2 minutes to log in and reach your publication dashboard
  await page.waitForTimeout(120000);

  // Save the session
  await context.storageState({ path: STATE_PATH });
  console.log('✅ Saved session to', STATE_PATH);

  await browser.close();
})();
