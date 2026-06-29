import { describe, it, expect } from 'vitest';
import { addDays, parseLocalDate } from './date.util';

describe('date util (timezone-safe YYYY-MM-DD handling)', () => {
  describe('addDays', () => {
    it('advances by one day', () => {
      expect(addDays('2026-06-29', 1)).toBe('2026-06-30');
    });

    it('rolls across a month boundary', () => {
      expect(addDays('2026-07-31', 1)).toBe('2026-08-01');
    });

    it('rolls across a year boundary', () => {
      expect(addDays('2026-12-31', 1)).toBe('2027-01-01');
    });

    it('goes backwards across a month boundary', () => {
      expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
    });

    it('returns the same date when adding zero', () => {
      expect(addDays('2026-06-29', 0)).toBe('2026-06-29');
    });
  });

  describe('parseLocalDate', () => {
    it('parses to local midnight on the same calendar day (no UTC shift)', () => {
      const d = parseLocalDate('2026-06-29');
      expect(d.getFullYear()).toBe(2026);
      expect(d.getMonth()).toBe(5); // June (0-based)
      expect(d.getDate()).toBe(29);
    });
  });
});
