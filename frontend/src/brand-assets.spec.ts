import { readFileSync } from 'node:fs';
import { join } from 'node:path';

/**
 * EL-071 — Soundcheck logo system (v1 draft) presence + validity guard.
 *
 * The brand SVGs under public/brand/ are the working assets the app reskin
 * (EL-078) consumes. This guard makes sure the set stays present and that each
 * file is a self-contained, valid single-root <svg> with a viewBox and no
 * external file references (which would break offline / silkscreen use).
 */
const BRAND_DIR = join(process.cwd(), 'public', 'brand');

const FILES = [
  'lockup-on-dark.svg',
  'mark.svg',
  'wordmark.svg',
  'one-color.svg',
  'favicon.svg',
  'favicon-16.svg',
  'status-cleared.svg',
  'status-pending.svg',
  'status-blocked.svg',
];

describe('Soundcheck brand assets (EL-071)', () => {
  it.each(FILES)('%s is a valid, self-contained single-root SVG', (name) => {
    const svg = readFileSync(join(BRAND_DIR, name), 'utf8');
    // single root <svg ... > ... </svg>
    expect(svg).toMatch(/<svg[\s>]/);
    expect(svg.match(/<svg[\s>]/g)?.length).toBe(1);
    expect(svg).toContain('</svg>');
    // viewBox present so it scales crisply
    expect(svg).toMatch(/viewBox="[^"]+"/);
    // no external refs (external images / xlink hrefs)
    expect(svg).not.toContain('xlink:href');
    expect(svg).not.toMatch(/href="https?:/);
  });

  it('one-color version carries no gradient (silkscreen-safe)', () => {
    const svg = readFileSync(join(BRAND_DIR, 'one-color.svg'), 'utf8');
    expect(svg).not.toContain('Gradient');
  });
});

/**
 * EL-077 — Soundcheck shell branding (favicon / PWA / OG / share-card) guard.
 *
 * The "chrome" surfaces outside the app UI must all read as Soundcheck: the
 * browser favicon, the installable PWA manifest + theme colour, the social OG
 * image, and the student promo-card attribution. This guard pins each wiring so
 * a future edit can't silently revert a shell to EventLineup.
 */
const BOOTH_BLACK = '#0B0D10';

describe('Soundcheck shell branding (EL-077)', () => {
  const index = readFileSync(join(process.cwd(), 'src', 'index.html'), 'utf8');

  it('index.html wires the SVG favicon (+16px variant), not the old .ico', () => {
    expect(index).toContain('href="brand/favicon.svg"');
    expect(index).toContain('href="brand/favicon-16.svg"');
    expect(index).not.toContain('favicon.ico');
  });

  it('index.html links the PWA manifest and sets the Booth Black theme colour', () => {
    expect(index).toContain('rel="manifest"');
    expect(index).toContain('href="manifest.webmanifest"');
    expect(index).toMatch(/<meta name="theme-color" content="#0B0D10">/);
  });

  it('manifest.webmanifest is valid Soundcheck JSON with Booth Black colours', () => {
    const raw = readFileSync(join(process.cwd(), 'public', 'manifest.webmanifest'), 'utf8');
    const manifest = JSON.parse(raw);
    expect(manifest.name).toBe('Soundcheck');
    expect(manifest.short_name).toBe('Soundcheck');
    expect(manifest.theme_color).toBe(BOOTH_BLACK);
    expect(manifest.background_color).toBe(BOOTH_BLACK);
    expect(Array.isArray(manifest.icons)).toBe(true);
    expect(manifest.icons.length).toBeGreaterThan(0);
    // Every manifest icon must point at a brand asset that actually exists.
    for (const icon of manifest.icons) {
      const rel = String(icon.src).replace(/^\//, '');
      expect(() => readFileSync(join(process.cwd(), 'public', rel), 'utf8')).not.toThrow();
    }
    // A maskable icon is present so the installed app icon crops cleanly.
    expect(manifest.icons.some((i: { purpose?: string }) => i.purpose === 'maskable')).toBe(true);
  });

  it('the maskable app icon is a full-bleed Booth tile (valid single-root SVG)', () => {
    const svg = readFileSync(join(BRAND_DIR, 'app-icon-maskable.svg'), 'utf8');
    expect(svg.match(/<svg[\s>]/g)?.length).toBe(1);
    expect(svg).toContain('</svg>');
    expect(svg).toMatch(/viewBox="[^"]+"/);
    expect(svg).toContain(BOOTH_BLACK);
  });

  it('the OG social image referenced by index.html exists and is a PNG', () => {
    expect(index).toContain('/assets/og/soundcheck-og.png');
    const png = readFileSync(join(process.cwd(), 'src', 'assets', 'og', 'soundcheck-og.png'));
    // PNG magic number.
    expect(png.subarray(0, 8).equals(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]))).toBe(true);
  });

  it('the student promo card is attributed to Soundcheck, not EventLineup', () => {
    const html = readFileSync(join(process.cwd(), 'src', 'app', 'card', 'card.component.html'), 'utf8');
    expect(html).toContain('MADE WITH SOUNDCHECK');
    expect(html).not.toMatch(/EVENTLINEUP/i);
  });
});
