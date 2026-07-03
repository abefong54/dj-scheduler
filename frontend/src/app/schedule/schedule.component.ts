import { Component, OnDestroy, signal, computed, effect } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot } from '../services/api.service';
import { LanguageService } from '../services/language.service';
import { slotDurationMins, toMinutes } from '../shared/slot-time.util';
import { addDays } from '../shared/date.util';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { ButtonComponent } from '../shared/button.component';
import { LoadingMeterComponent } from '../shared/loading-meter.component';

@Component({
  selector: 'app-schedule',
  standalone: true,
  imports: [TranslatePipe, ButtonComponent, LoadingMeterComponent],
  templateUrl: './schedule.component.html',
  styleUrl: './schedule.component.css',
})
export class ScheduleComponent implements OnDestroy {
  private destroy$ = new Subject<void>();

  event = signal<Event | null>(null);
  stages = signal<Stage[]>([]);
  slots = signal<Slot[]>([]);
  selectedDate = signal<string>('');
  timeHeaders = signal<string[]>([]);
  loading = signal(true);
  loadError = signal(false);
  private eventId = '';

  dates = computed(() => {
    const e = this.event();
    if (!e) return [];
    return this.dateRange(e.start_date, e.end_date);
  });

  slotsForDate = computed(() => {
    return this.slots().filter(s => s.slot_date === this.selectedDate());
  });

  stagesForDate = computed(() => {
    const ids = new Set(this.slotsForDate().map(s => s.stage_id));
    return this.stages().filter(s => ids.has(s.id));
  });

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    public langService: LanguageService,
  ) {
    this.eventId = this.route.snapshot.paramMap.get('id')!;
    this.load();

    effect(() => {
      this.slots();
      this.updateTimeHeaders();
    });
  }

  // This page is public (no auth guard), so it reads the unauthenticated /public
  // endpoint — the protected endpoints would 401 a visitor and bounce them to
  // /login (EL-034).
  load() {
    this.loading.set(true);
    this.loadError.set(false);
    this.api.getPublicSchedule(this.eventId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: ({ event, stages, slots }) => {
          this.event.set(event);
          this.stages.set(stages);
          this.slots.set(slots);
          const newDates = this.dateRange(event.start_date, event.end_date);
          if (newDates.length > 0 && this.selectedDate() === '') {
            this.selectedDate.set(newDates[0]);
          }
          this.updateTimeHeaders();
          this.loading.set(false);
        },
        error: () => {
          this.loadError.set(true);
          this.loading.set(false);
        },
      });
  }

  retry() {
    this.load();
  }

  slotsForStage(stageId: string): Slot[] {
    return this.slotsForDate()
      .filter(s => s.stage_id === stageId)
      .sort((a, b) => a.start_time.localeCompare(b.start_time));
  }

  slotsAt(stageId: string, time: string): Slot[] {
    const t = toMinutes(time);
    return this.slotsForDate().filter(s => {
      if (s.stage_id !== stageId) return false;
      const start = toMinutes(s.start_time);
      // Duration is wrap-aware, so a slot running past midnight still fills the
      // cells from its start to the end of the day's grid.
      return t >= start && t < start + slotDurationMins(s.start_time, s.end_time);
    });
  }

  stageColor(stageId: string): string {
    return this.stages().find(s => s.id === stageId)?.color ?? '#6366F1';
  }

  private updateTimeHeaders() {
    const slotsArray = this.slots();
    if (slotsArray.length === 0) {
      this.timeHeaders.set(this.generateHeaders('14:00', '23:00'));
      return;
    }
    const startMins = slotsArray.map(s => toMinutes(s.start_time));
    // End each slot on an absolute timeline so a past-midnight slot extends the
    // range upward; cap at 24:00 so the headers stop at the end of the day.
    const endMins = slotsArray.map(s =>
      Math.min(toMinutes(s.start_time) + slotDurationMins(s.start_time, s.end_time), 24 * 60));
    const startMin = Math.min(...startMins) - 30;
    const endMin = Math.max(...endMins);
    this.timeHeaders.set(
      this.generateHeadersFromMinutes(Math.max(startMin, 0), endMin)
    );
  }


  private generateHeaders(start: string, end: string): string[] {
    return this.generateHeadersFromMinutes(toMinutes(start), toMinutes(end));
  }

  private generateHeadersFromMinutes(startMin: number, endMin: number): string[] {
    const headers: string[] = [];
    let cur = Math.floor(startMin / 30) * 30;
    while (cur < endMin) {
      const h = Math.floor(cur / 60);
      const m = cur % 60;
      headers.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
      cur += 30;
    }
    return headers;
  }

  private dateRange(start: string, end: string): string[] {
    // Lexicographic comparison is correct for zero-padded YYYY-MM-DD, and addDays
    // steps the date without ever touching the local timezone (see date.util).
    const dates: string[] = [];
    let cur = start;
    while (cur <= end) {
      dates.push(cur);
      cur = addDays(cur, 1);
    }
    return dates;
  }

  setLang(lang: string) {
    this.langService.setLanguage(lang);
  }

  close() {
    window.close();
  }

  shareViaLine() {
    const url = encodeURIComponent(window.location.href);
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
