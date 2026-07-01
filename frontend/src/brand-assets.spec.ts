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
