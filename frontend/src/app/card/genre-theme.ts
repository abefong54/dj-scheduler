// Genre → card theme mapping for the student promo card (EL-052). IMDJ's real
// posters re-colour by genre, so the card auto-themes its accent + panel palette
// off the slot's genre. Unknown/empty genres fall back to the brand violet/lime
// default. The actual palettes live as CSS-variable overrides keyed on
// `.card.theme-<key>` in card.component.css.

export type GenreTheme = 'house' | 'techno' | 'hiphop' | 'soul' | 'default';

// Keyed by the genre lowercased with all non-letters stripped, so "Hip-Hop",
// "hip hop" and "HipHop" all collapse to "hiphop". Aliases group sub-genres and
// adjacent styles onto one of the four themes.
const GENRE_THEMES: Record<string, GenreTheme> = {
  // House / four-on-the-floor → green
  house: 'house',
  deephouse: 'house',
  techhouse: 'house',
  afrohouse: 'house',
  progressivehouse: 'house',
  // Techno / electronic → steel cyan
  techno: 'techno',
  minimaltechno: 'techno',
  tech: 'techno',
  trance: 'techno',
  electro: 'techno',
  // Hip-Hop / rap → violet + orange
  hiphop: 'hiphop',
  rap: 'hiphop',
  trap: 'hiphop',
  drill: 'hiphop',
  // Soul / R&B / funk → blue + lime
  soul: 'soul',
  neosoul: 'soul',
  rnb: 'soul',
  rb: 'soul',
  randb: 'soul',
  funk: 'soul',
  disco: 'soul',
};

/** Maps a free-text genre to its card theme, defaulting to the brand palette. */
export function genreTheme(genre: string | null | undefined): GenreTheme {
  const key = (genre ?? '').toLowerCase().replace(/[^a-z]/g, '');
  return GENRE_THEMES[key] ?? 'default';
}
