import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * CSP frame-src guard — the Challenges feature embeds media players (YouTube,
 * SoundCloud, Mixcloud) in iframes. The app's authoritative Content-Security-
 * Policy lives in public/_headers (host layer, Netlify). With default-src
 * 'self', frames are blocked unless frame-src lists the providers, so this guard
 * pins the allowed embed hosts and makes sure frame-ancestors stays locked.
 */
// Anchor the fixture read to this spec's own location (src/), not process.cwd()
// — the Angular vitest builder resolves cwd inconsistently (EL-097).
const FRONTEND_ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const HEADERS = readFileSync(join(FRONTEND_ROOT, 'public', '_headers'), 'utf8');

const EMBED_HOSTS = [
  'https://www.youtube.com',
  'https://www.youtube-nocookie.com',
  'https://w.soundcloud.com',
  'https://soundcloud.com',
  'https://www.mixcloud.com',
  'https://player.mixcloud.com',
];

describe('CSP frame-src for challenge embeds', () => {
  // Pull the single Content-Security-Policy line out of the _headers file.
  // Match the actual header line ("  Content-Security-Policy: …"), not the
  // explanatory comment above it ("# … Content-Security-Policy …").
  const cspLine = HEADERS.split('\n').find((l) => l.trimStart().startsWith('Content-Security-Policy:')) ?? '';
  const frameSrc = /frame-src ([^;]+);/.exec(cspLine)?.[1] ?? '';

  it('declares a frame-src directive', () => {
    expect(cspLine).toContain('frame-src');
    expect(frameSrc.trim().length).toBeGreaterThan(0);
  });

  it.each(EMBED_HOSTS)('allows %s as a frame source', (host) => {
    expect(frameSrc.split(/\s+/)).toContain(host);
  });

  it('keeps frame-ancestors locked to none (we embed others, not vice versa)', () => {
    expect(cspLine).toContain("frame-ancestors 'none'");
  });
});
