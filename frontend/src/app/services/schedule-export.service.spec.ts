import * as XLSX from 'xlsx-js-style';
import { ScheduleExportService, ExportLabels, ExportSlot } from './schedule-export.service';

const EN_LABELS: ExportLabels = {
  date: 'Date',
  stage: 'Stage',
  timeSlot: 'Time Slot',
  dj: 'DJ',
  genre: 'Genre',
  notes: 'Notes',
};

function slot(partial: Partial<ExportSlot>): ExportSlot {
  return {
    slot_date: '2026-07-01',
    stage_name: 'Main',
    dj_name: 'DJ A',
    start_time: '16:00',
    end_time: '17:00',
    genre: '',
    notes: '',
    ...partial,
  };
}

describe('ScheduleExportService', () => {
  const svc = new ScheduleExportService();

  describe('formatTimeSlot', () => {
    it('formats a 90-minute slot as fractional hours', () => {
      expect(svc.formatTimeSlot('16:00', '17:30')).toBe('16:00 - 17:30 (1.5hr)');
    });

    it('formats a whole-hour slot without a decimal', () => {
      expect(svc.formatTimeSlot('18:00', '19:00')).toBe('18:00 - 19:00 (1hr)');
    });

    it('formats a two-hour slot', () => {
      expect(svc.formatTimeSlot('20:00', '22:00')).toBe('20:00 - 22:00 (2hr)');
    });

    it('formats a half-hour slot', () => {
      expect(svc.formatTimeSlot('16:00', '16:30')).toBe('16:00 - 16:30 (0.5hr)');
    });
  });

  describe('filename', () => {
    it('appends -schedule.xlsx and replaces spaces with dashes', () => {
      expect(svc.filename('Summer Bash 2026')).toBe('Summer-Bash-2026-schedule.xlsx');
    });

    it('collapses runs of whitespace into single dashes', () => {
      expect(svc.filename('  My   Event ')).toBe('My-Event-schedule.xlsx');
    });
  });

  describe('buildAoa', () => {
    it('puts the event name in row 1 and headers in row 2', () => {
      const aoa = svc.buildAoa('My Event', [slot({})], EN_LABELS);
      expect(aoa[0]).toEqual(['My Event']);
      expect(aoa[1]).toEqual(['Date', 'Stage', 'Time Slot', 'DJ', 'Genre', 'Notes']);
    });

    it('emits a data row with the six columns in order', () => {
      const aoa = svc.buildAoa('E', [
        slot({ slot_date: '2026-07-01', stage_name: 'Main', dj_name: 'Marcus', start_time: '16:00', end_time: '17:30', genre: 'Techno', notes: 'mic check' }),
      ], EN_LABELS);
      expect(aoa[2]).toEqual(['2026-07-01', 'Main', '16:00 - 17:30 (1.5hr)', 'Marcus', 'Techno', 'mic check']);
    });

    it('sorts by date, then stage name, then start time', () => {
      const aoa = svc.buildAoa('E', [
        slot({ slot_date: '2026-07-02', stage_name: 'Main', start_time: '10:00', dj_name: 'D2' }),
        slot({ slot_date: '2026-07-01', stage_name: 'Patio', start_time: '09:00', dj_name: 'D1b' }),
        slot({ slot_date: '2026-07-01', stage_name: 'Main', start_time: '12:00', dj_name: 'D1a-late' }),
        slot({ slot_date: '2026-07-01', stage_name: 'Main', start_time: '09:00', dj_name: 'D1a-early' }),
      ], EN_LABELS);
      // Skip title+header; first data block is 2026-07-01, Main(09,12) then Patio(09)
      const djCol = aoa.slice(2).map(r => r[3]);
      expect(djCol).toEqual(['D1a-early', 'D1a-late', 'D1b', '', 'D2']);
    });

    it('inserts one blank row between date groups but not at the ends', () => {
      const aoa = svc.buildAoa('E', [
        slot({ slot_date: '2026-07-01', dj_name: 'A' }),
        slot({ slot_date: '2026-07-02', dj_name: 'B' }),
      ], EN_LABELS);
      // row0 title, row1 headers, row2 A, row3 blank, row4 B
      expect(aoa.length).toBe(5);
      expect(aoa[3]).toEqual(['', '', '', '', '', '']);
      expect(aoa[2][3]).toBe('A');
      expect(aoa[4][3]).toBe('B');
    });

    it('renders an empty Genre cell when genre is undefined', () => {
      const s = slot({ genre: undefined });
      const aoa = svc.buildAoa('E', [s], EN_LABELS);
      expect(aoa[2][4]).toBe('');
    });

    it('uses the provided (translated) headers', () => {
      const zh: ExportLabels = { date: '日期', stage: '舞台', timeSlot: '時段', dj: 'DJ', genre: '風格', notes: '備註' };
      const aoa = svc.buildAoa('E', [slot({})], zh);
      expect(aoa[1]).toEqual(['日期', '舞台', '時段', 'DJ', '風格', '備註']);
    });
  });

  describe('buildWorkbook', () => {
    const SAMPLE = [
      slot({ slot_date: '2026-07-01', stage_name: 'Main', dj_name: 'Marcus', start_time: '16:00', end_time: '17:30', genre: 'Techno', notes: 'mic' }),
    ];

    it('sets a bold title and merges it across all six columns', () => {
      const wb = svc.buildWorkbook('Summer Bash', SAMPLE, EN_LABELS);
      const ws = wb.Sheets[wb.SheetNames[0]];
      expect(ws['A1'].v).toBe('Summer Bash');
      expect(ws['A1'].s.font.bold).toBe(true);
      expect(ws['!merges']).toEqual([{ s: { r: 0, c: 0 }, e: { r: 0, c: 5 } }]);
      expect(ws['A2'].s.font.bold).toBe(true); // header row bold too
    });

    it('serializes to valid .xlsx bytes that read back with the right values', () => {
      const wb = svc.buildWorkbook('Summer Bash', SAMPLE, EN_LABELS);
      const buf = XLSX.write(wb, { type: 'array', bookType: 'xlsx' });
      const ws = XLSX.read(buf, { type: 'array' }).Sheets['Schedule'];
      expect(ws['A1'].v).toBe('Summer Bash');
      expect(ws['!merges']).toEqual([{ s: { r: 0, c: 0 }, e: { r: 0, c: 5 } }]);
      expect([ws['A2'].v, ws['C2'].v, ws['F2'].v]).toEqual(['Date', 'Time Slot', 'Notes']);
      expect([ws['A3'].v, ws['C3'].v, ws['D3'].v, ws['E3'].v]).toEqual(['2026-07-01', '16:00 - 17:30 (1.5hr)', 'Marcus', 'Techno']);
    });
  });
});
