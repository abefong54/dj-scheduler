import { describe, it, expect } from 'vitest';
import { slotDurationMins, spansNextDay, toMinutes } from './slot-time.util';

describe('slot-time util (end <= start means the slot runs into the next day)', () => {
  describe('toMinutes', () => {
    it('converts HH:MM to minutes since midnight', () => {
      expect(toMinutes('00:00')).toBe(0);
      expect(toMinutes('23:30')).toBe(1410);
    });
  });

  describe('slotDurationMins', () => {
    it('returns the plain difference for a same-day slot', () => {
      expect(slotDurationMins('18:00', '19:30')).toBe(90);
    });

    it('returns a positive duration for a slot that crosses midnight', () => {
      // 23:30 -> 00:30 is one hour, not minus twenty-three.
      expect(slotDurationMins('23:30', '00:30')).toBe(60);
    });

    it('treats an end of exactly midnight as ending the day (not zero / not a full day)', () => {
      expect(slotDurationMins('23:00', '00:00')).toBe(60);
    });
  });

  describe('spansNextDay', () => {
    it('is false for a same-day slot', () => {
      expect(spansNextDay('18:00', '19:00')).toBe(false);
    });

    it('is true when the end is at or before the start', () => {
      expect(spansNextDay('23:30', '00:30')).toBe(true);
      expect(spansNextDay('23:00', '00:00')).toBe(true);
    });
  });
});
