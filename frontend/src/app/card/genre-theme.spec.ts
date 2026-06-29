import { genreTheme } from './genre-theme';

describe('genreTheme (EL-052)', () => {
  it('maps the four core genres to their themes', () => {
    expect(genreTheme('House')).toBe('house');
    expect(genreTheme('Techno')).toBe('techno');
    expect(genreTheme('Hip-Hop')).toBe('hiphop');
    expect(genreTheme('Soul')).toBe('soul');
  });

  it('normalizes case, spaces and punctuation', () => {
    expect(genreTheme('hip hop')).toBe('hiphop');
    expect(genreTheme('HIP-HOP')).toBe('hiphop');
    expect(genreTheme('Deep House')).toBe('house');
    expect(genreTheme('Tech House')).toBe('house');
  });

  it('groups adjacent styles via aliases', () => {
    expect(genreTheme('R&B')).toBe('soul'); // blue/lime, per IMDJ posters
    expect(genreTheme('RnB')).toBe('soul');
    expect(genreTheme('Funk')).toBe('soul');
    expect(genreTheme('Trap')).toBe('hiphop');
    expect(genreTheme('Trance')).toBe('techno');
  });

  it('falls back to the brand default for unknown or empty genres', () => {
    expect(genreTheme('Reggae')).toBe('default');
    expect(genreTheme('')).toBe('default');
    expect(genreTheme(null)).toBe('default');
    expect(genreTheme(undefined)).toBe('default');
  });
});
