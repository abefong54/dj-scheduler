import { readFileSync } from 'node:fs';
import { join } from 'node:path';

/**
 * EL-073 — Custom domain icon set guard.
 *
 * The branded domain glyphs (waveform-check, half-meter, fader, booth,
 * certificate) are shipped as static SVG asset files under
 * `public/brand/icons/`. They are not yet wired into any Angular component
 * (the per-surface reskin tickets do that), so a rendering test is not
 * meaningful here. Instead we guard the invariants that make them safe,
 * reusable assets: each file must exist and be a single-root <svg> with a
 * viewBox and no external references. Mirrors the source-of-truth style of
 * `styles-tokens.spec.ts`.
 *
 * Vitest runs with the frontend package dir as cwd.
 */
const ICONS_DIR = join(process.cwd(), 'public', 'brand', 'icons');

const ICON_FILES = [
  'waveform-check-line.svg',
  'waveform-check-filled.svg',
  'half-meter-line.svg',
  'half-meter-filled.svg',
  'fader-line.svg',
  'fader-filled.svg',
  'booth-line.svg',
  'booth-filled.svg',
  'certificate-line.svg',
  'certificate-filled.svg',
];

describe('EventLineup domain icons (EL-073)', () => {
  it.each(ICON_FILES)('%s exists and parses as a single-root <svg> with a viewBox', (file) => {
    const svg = readFileSync(join(ICONS_DIR, file), 'utf8');

    // Exactly one root <svg> element.
    const openTags = svg.match(/<svg[\s>]/g) ?? [];
    const closeTags = svg.match(/<\/svg>/g) ?? [];
    expect(openTags.length, `${file}: exactly one <svg> open tag`).toBe(1);
    expect(closeTags.length, `${file}: exactly one </svg> close tag`).toBe(1);

    // The document must start with the <svg> root (allowing a leading BOM/space
    // or an <?xml?> declaration), so there is a single root element.
    expect(svg.replace(/^﻿/, '').trimStart(), `${file}: starts at the <svg> root`).toMatch(
      /^(?:<\?xml[^>]*\?>\s*)?<svg[\s>]/,
    );

    // A 24px viewBox so the glyph scales cleanly at 16px and 24px.
    expect(svg, `${file}: declares a 24-unit viewBox`).toContain('viewBox="0 0 24 24"');

    // No external references — these must be self-contained assets.
    expect(svg, `${file}: no external url() ref`).not.toMatch(/url\(\s*['"]?https?:/i);
    expect(svg, `${file}: no <image>/<use> external ref`).not.toMatch(/<(image|use)\b/i);
    expect(svg, `${file}: no external xlink:href`).not.toMatch(/xlink:href/i);
  });

  it('status-semantic filled glyphs bake their token color; line glyphs stay currentColor', () => {
    // Filled status glyphs carry meaning through color (paired with a label).
    expect(readFileSync(join(ICONS_DIR, 'waveform-check-filled.svg'), 'utf8')).toContain('#34D399');
    expect(readFileSync(join(ICONS_DIR, 'half-meter-filled.svg'), 'utf8')).toContain('#F5B544');

    // Non-status glyphs inherit the token color via currentColor.
    for (const file of ['fader-line.svg', 'booth-line.svg', 'certificate-line.svg']) {
      expect(readFileSync(join(ICONS_DIR, file), 'utf8'), `${file}: uses currentColor`).toContain(
        'currentColor',
      );
    }
  });
});
