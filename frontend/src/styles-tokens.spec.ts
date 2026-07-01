import { readFileSync } from 'node:fs';
import { join } from 'node:path';

/**
 * EL-072 — Soundcheck design tokens (source-of-truth) regression guard.
 *
 * These tokens are the brand SoT, so the exact hex values matter and must not
 * drift silently. They are also an ADDITIVE, NON-BREAKING layer: they must stay
 * opt-in (scoped under `[data-brand="soundcheck"]`) and must never leak into
 * `:root`, where they would override the live Console defaults.
 *
 * A getComputedStyle/DOM assertion is not meaningful here: the global
 * `styles.css` is not loaded into the jsdom unit-test bed, and jsdom does not
 * resolve author-stylesheet custom properties. So we assert against the source
 * of truth directly — which is exactly what this ticket is about.
 */
// Vitest runs with the frontend package dir as cwd; styles.css is the global
// SoT stylesheet under src/.
const css = readFileSync(join(process.cwd(), 'src', 'styles.css'), 'utf8');

describe('Soundcheck design tokens (EL-072)', () => {
  it('defines the brand accent + status tokens with the spec hex values', () => {
    const expected: Record<string, string> = {
      '--sc-cue-amber': '#ff9e2c',
      '--sc-deck-cyan': '#22d3ee',
      '--sc-foglight-magenta': '#e0469b',
      '--sc-booth-black': '#0b0d10',
      '--sc-console-slate': '#171b21',
      '--sc-stage-white': '#f7f8fa',
      '--sc-status-cleared': '#34d399',
      '--sc-status-pending': '#f5b544',
      '--sc-status-blocked': '#f4505b',
      '--sc-status-info': '#22d3ee',
    };
    for (const [token, hex] of Object.entries(expected)) {
      expect(css).toContain(`${token}: ${hex};`);
    }
  });

  it('defines the full dark-mode neutral ramp', () => {
    expect(css).toContain('--sc-neutral-950: #0b0d10;');
    expect(css).toContain('--sc-neutral-900: #171b21;');
    expect(css).toContain('--sc-neutral-700: #2c333c;');
    expect(css).toContain('--sc-neutral-400: #8a94a3;');
    expect(css).toContain('--sc-neutral-100: #e6eaf0;');
  });

  it('wires Noto Sans TC into the sans + display type stacks (CJK coverage)', () => {
    expect(css).toContain(
      "--sc-font-display: 'Space Grotesk', 'Noto Sans TC', system-ui, sans-serif;",
    );
    expect(css).toContain(
      "--sc-font-body: 'Inter', 'Noto Sans TC', system-ui, sans-serif;",
    );
  });

  it('ships both a booth (dark) and an admin (light) mode selector', () => {
    expect(css).toContain("[data-brand='soundcheck'] {");
    expect(css).toContain("[data-brand='soundcheck'][data-mode='light'] {");
  });

  it('is opt-in and non-breaking: no --sc-* token is declared in :root', () => {
    // Isolate the :root { ... } block (Console tokens live here) and assert it
    // carries no Soundcheck token, so the live look is untouched by default.
    const rootBlock = css.slice(css.indexOf(':root {'));
    const rootBody = rootBlock.slice(0, rootBlock.indexOf('}'));
    expect(rootBody).not.toContain('--sc-');
  });
});
