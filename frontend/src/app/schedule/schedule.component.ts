import { Component, OnDestroy, signal, computed, effect } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot } from '../services/api.service';
import { LanguageService } from '../services/language.service';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

@Component({
  selector: 'app-schedule',
  standalone: true,
  imports: [TranslatePipe],
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
    const eventId = this.route.snapshot.paramMap.get('id')!;

    this.api.getEvent(eventId)
      .pipe(takeUntil(this.destroy$))
      .subscribe(e => {
        this.event.set(e);
        const newDates = this.dateRange(e.start_date, e.end_date);
        if (newDates.length > 0 && this.selectedDate() === '') {
          this.selectedDate.set(newDates[0]);
        }
      });

    this.api.getStages(eventId)
      .pipe(takeUntil(this.destroy$))
      .subscribe(s => this.stages.set(s));

    this.api.getSlots(eventId)
      .pipe(takeUntil(this.destroy$))
      .subscribe(s => {
        this.slots.set(s);
        this.updateTimeHeaders();
      });

    effect(() => {
      this.slots();
      this.updateTimeHeaders();
    });
  }

  slotsForStage(stageId: string): Slot[] {
    return this.slotsForDate()
      .filter(s => s.stage_id === stageId)
      .sort((a, b) => a.start_time.localeCompare(b.start_time));
  }

  slotsAt(stageId: string, time: string): Slot[] {
    return this.slotsForDate().filter(s =>
      s.stage_id === stageId && s.start_time <= time && s.end_time > time);
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
    const starts = slotsArray.map(s => s.start_time).sort();
    const ends = slotsArray.map(s => s.end_time).sort();
    const earliest = starts[0];
    const latest = ends[ends.length - 1];
    const startMin = this.toMinutes(earliest) - 30;
    const endMin = this.toMinutes(latest);
    this.timeHeaders.set(
      this.generateHeadersFromMinutes(Math.max(startMin, 0), endMin)
    );
  }

  private toMinutes(t: string): number {
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  private generateHeaders(start: string, end: string): string[] {
    return this.generateHeadersFromMinutes(
      this.toMinutes(start), this.toMinutes(end));
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
    const dates: string[] = [];
    const cur = new Date(start);
    const last = new Date(end);
    while (cur <= last) {
      dates.push(cur.toISOString().slice(0, 10));
      cur.setDate(cur.getDate() + 1);
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
