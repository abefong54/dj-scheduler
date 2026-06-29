// Calendar-date helpers for "YYYY-MM-DD" strings.
//
// `new Date('2026-06-29')` parses as UTC midnight, so iterating it and reading
// it back with toISOString() can land on the wrong day for viewers west of UTC.
// These helpers keep date math off the local timezone entirely.

/** Add (or subtract) whole days to a "YYYY-MM-DD" string, timezone-independent. */
export function addDays(date: string, days: number): string {
  const [y, m, d] = date.split('-').map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  dt.setUTCDate(dt.getUTCDate() + days);
  return dt.toISOString().slice(0, 10);
}

/**
 * Parse a "YYYY-MM-DD" string to a Date at *local* midnight, so formatting it
 * with toLocaleDateString shows the same calendar day the string names.
 */
export function parseLocalDate(date: string): Date {
  const [y, m, d] = date.split('-').map(Number);
  return new Date(y, m - 1, d);
}
