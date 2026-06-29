// Slot times are stored as a slot_date plus a start_time and end_time (HH:MM).
// A slot whose end_time is at or before its start_time runs past midnight into
// the next day — e.g. a 23:30 set ending 00:30. These helpers keep that
// convention in one place so every consumer (duration display, conflict
// detection, export, grid) interprets a wrapped slot the same way.

const MINUTES_PER_DAY = 24 * 60;

/** Minutes since midnight for an "HH:MM" (or "HH:MM:SS") time string. */
export function toMinutes(time: string): number {
  const [h, m] = time.split(':').map(Number);
  return h * 60 + m;
}

/** True when the slot ends on the following day (end is at or before start). */
export function spansNextDay(start: string, end: string): boolean {
  return toMinutes(end) <= toMinutes(start);
}

/** Slot length in minutes, accounting for a slot that runs into the next day. */
export function slotDurationMins(start: string, end: string): number {
  const diff = toMinutes(end) - toMinutes(start);
  return diff > 0 ? diff : diff + MINUTES_PER_DAY;
}
