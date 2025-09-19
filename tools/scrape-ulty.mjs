import fs from 'fs';
import path from 'path';
import https from 'https';
import { fileURLToPath } from 'url';
import puppeteer from 'puppeteer';
import csvParser from 'csv-parser';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function downloadCSV(url, outputPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(outputPath);
    https.get(url, (response) => {
      if (response.statusCode !== 200) return reject(new Error(`HTTP ${response.statusCode}`));
      response.pipe(file);
      file.on('finish', () => file.close(resolve));
    }).on('error', (err) => { fs.unlink(outputPath, () => reject(err)); });
  });
}

async function extractDataFromCSV(filePath) {
  const rows = [];
  return new Promise((resolve, reject) => {
    fs.createReadStream(filePath)
      .pipe(csvParser())
      .on('data', (row) => rows.push(row))
      .on('end', () => {
        if (rows.length === 0) return reject(new Error('Empty CSV'));
        const headers = Object.keys(rows[0]);
        const dateCol = headers.find(h => h.toLowerCase().includes('date'));
        const ticCol  = headers.find(h => h.toLowerCase().includes('ticker'));
        if (!dateCol || !ticCol) return reject(new Error('Missing Date/Ticker columns'));
        const seen = new Set();
        let latestDate = null;
        for (const r of rows) {
          const d = r[dateCol];
          const t = r[ticCol];
          if (t && !/\s+\S+/.test(t.trim())) seen.add(t.trim());
          if (d && (!latestDate || new Date(d) > new Date(latestDate))) latestDate = d;
        }
        resolve({ latestDate, stockTickers: [...seen] });
      })
      .on('error', reject);
  });
}

async function main() {
  const outDir = path.resolve(__dirname, '..', 'data'); // always repo-level
  const tmpCsv = path.join(process.cwd(), 'holdings.csv');
  fs.mkdirSync(outDir, { recursive: true });

  const browser = await puppeteer.launch({ headless: 'new', args: ['--no-sandbox','--disable-setuid-sandbox'] });
  try {
    const page = await browser.newPage();
    await page.goto('https://www.yieldmaxetfs.com/ulty/', { waitUntil: 'networkidle0', timeout: 120000 });

    const downloadUrl = await page.evaluate(() => {
      const link = [...document.querySelectorAll('a')].find(a => a.textContent.trim() === 'Download All Holdings');
      return link ? link.href : null;
    });
    if (!downloadUrl) throw new Error('Download link not found');

    await downloadCSV(downloadUrl, tmpCsv);
    const { latestDate, stockTickers } = await extractDataFromCSV(tmpCsv);

    const dividendYield = await page.evaluate(() => {
      const cell = document.querySelector('td.column-rate');
      return cell ? cell.textContent.trim() : null;
    });

    const payload = { latestDate, stockTickers, dividendYield: dividendYield || 'Not Found' };
    fs.writeFileSync(path.join(outDir, 'rotations.json'), JSON.stringify(payload, null, 2));
    console.log(`Wrote data/rotations.json with ${payload.stockTickers.length} tickers on ${latestDate}`);
  } finally {
    await browser.close().catch(()=>{});
    if (fs.existsSync(tmpCsv)) fs.unlinkSync(tmpCsv);
  }
}

main().catch((e) => { console.error(e); process.exit(1); });
