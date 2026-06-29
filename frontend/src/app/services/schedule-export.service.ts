import { Injectable } from '@angular/core';
import * as XLSX from 'xlsx-js-style';
import { slotDurationMins } from '../shared/slot-time.util';

/** Column header strings, already translated to the active language. */
export interface ExportLabels {
  date: string;
  stage: string;
  timeSlot: string;
  dj: string;
  genre: string;
  notes: string;
}

/**
 * The slot shape the export needs. Structurally a subset of api.service `Slot`,
 * with `genre` optional because it is not yet wired through the frontend model.
 */
export interface ExportSlot {
  slot_date: string;   // 'YYYY-MM-DD'
  stage_name: string;
  dj_name: string;
  start_time: string;  // 'HH:MM'
  end_time: string;    // 'HH:MM'
  notes: string;
  genre?: string;
}

const COLUMN_COUNT = 6;
const EMPTY_ROW = ['', '', '', '', '', ''];

@Injectable({ providedIn: 'root' })
export class ScheduleExportService {
  /** "16:00 - 17:30 (1.5hr)" — whole hours render without a decimal. A slot that
   *  runs past midnight (end <= start) counts as next-day, so its hours stay positive. */
  formatTimeSlot(start: string, end: string): string {
    const hours = slotDurationMins(start, end) / 60;
    const rounded = parseFloat(hours.toFixed(2));
    return `${start} - ${end} (${rounded}hr)`;
  }

  /** "Summer Bash" -> "Summer-Bash-schedule.xlsx" */
  filename(eventName: string): string {
    const slug = eventName.trim().replace(/\s+/g, '-');
    return `${slug}-schedule.xlsx`;
  }

  /**
   * Build the array-of-arrays for the sheet: title row, header row, then data
   * rows sorted by date -> stage -> start time, with a blank row between date
   * groups. Pure and deterministic so it can be unit-tested without the DOM.
   */
  buildAoa(eventName: string, slots: ExportSlot[], labels: ExportLabels): string[][] {
    const sorted = [...slots].sort(
      (a, b) =>
        a.slot_date.localeCompare(b.slot_date) ||
        a.stage_name.localeCompare(b.stage_name) ||
        a.start_time.localeCompare(b.start_time),
    );

    const rows: string[][] = [
      [eventName],
      [labels.date, labels.stage, labels.timeSlot, labels.dj, labels.genre, labels.notes],
    ];

    let prevDate: string | null = null;
    for (const s of sorted) {
      if (prevDate !== null && s.slot_date !== prevDate) {
        rows.push([...EMPTY_ROW]);
      }
      rows.push([
        s.slot_date,
        s.stage_name,
        this.formatTimeSlot(s.start_time, s.end_time),
        s.dj_name,
        s.genre ?? '',
        s.notes ?? '',
      ]);
      prevDate = s.slot_date;
    }

    return rows;
  }

  /**
   * Build the styled workbook (title bold + merged, header row bold). Separated
   * from the download so the .xlsx output can be round-tripped in tests.
   */
  buildWorkbook(eventName: string, slots: ExportSlot[], labels: ExportLabels): XLSX.WorkBook {
    const aoa = this.buildAoa(eventName, slots, labels);
    const ws = XLSX.utils.aoa_to_sheet(aoa);

    // Title merged across all columns, bold.
    ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: COLUMN_COUNT - 1 } }];
    const titleStyle = { font: { bold: true, sz: 14 } };
    const headerStyle = { font: { bold: true } };
    if (ws['A1']) ws['A1'].s = titleStyle;
    for (let c = 0; c < COLUMN_COUNT; c++) {
      const ref = XLSX.utils.encode_cell({ r: 1, c });
      if (ws[ref]) ws[ref].s = headerStyle;
    }

    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Schedule');
    return wb;
  }

  /** Build the workbook and trigger a client-side .xlsx download. */
  download(eventName: string, slots: ExportSlot[], labels: ExportLabels): void {
    const wb = this.buildWorkbook(eventName, slots, labels);
    XLSX.writeFile(wb, this.filename(eventName));
  }
}
