import { writeFile } from 'node:fs/promises';

const SITE_URL = (process.env.SITE_URL || 'https://gorenel.site').replace(/\/+$/, '');

// Keep this as the single source of truth for indexable public URLs.
const URLS = [
  '/',
  '/tr/ngrok-alternatifi',
  '/tr/localhost-disari-acma',
  '/tr/webhook-local-test',
  '/tr/docs/cli',
  '/tr/gizlilik-politikasi',
  '/en/privacy',
  '/tr/docs/api',
  '/en/docs/api',
];

function xmlEscape(s) {
  return s
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;');
}

const entries = URLS.map((path) => {
  const loc = `${SITE_URL}${path === '/' ? '' : path}`;
  return `  <url>\n    <loc>${xmlEscape(loc)}</loc>\n  </url>`;
}).join('\n');

const sitemap = `<?xml version="1.0" encoding="UTF-8"?>\n` +
  `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n` +
  `${entries}\n` +
  `</urlset>\n`;

await writeFile(new URL('../public/sitemap.xml', import.meta.url), sitemap, 'utf8');
console.log(`[sitemap] wrote ${URLS.length} urls to public/sitemap.xml`);

