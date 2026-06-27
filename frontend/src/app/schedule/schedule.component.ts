// frontend/src/app/schedule/schedule.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot } from '../services/api.service';

@Component({
  selector: 'app-schedule',
  standalone: true,
  imports: [CommonModule, TranslatePipe],
  template: `
    <div class="min-h-dvh bg-gray-50">
      <!-- Public header (no nav) -->
      <div class="bg-white border-b border-gray-200 px-4 py-4 sticky top-0 z-10">
        <div class="max-w-5xl mx-auto flex items-start justify-between gap-4">
          <div *ngIf="event">
            <h1 class="text-lg font-bold text-gray-900 leading-tight">{{ event.name }}</h1>
            <p class="text-sm text-gray-500 mt-0.5">{{ event.venue_name }}</p>
          </div>
          <div class="flex gap-1 shrink-0">
            <button (click)="setLang('en')"
              class="px-3 py-1 rounded text-xs font-medium transition-colors"
              [class]="currentLang==='en'
                ? 'bg-slate-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'">EN</button>
            <button (click)="setLang('zh-TW')"
              class="px-3 py-1 rounded text-xs font-medium transition-colors"
              [class]="currentLang==='zh-TW'
                ? 'bg-slate-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'">中文</button>
          </div>
        </div>
      </div>

      <div class="max-w-5xl mx-auto px-4 py-4">
        <!-- LINE share button -->
        <button (click)="shareViaLine()"
          class="w-full sm:w-auto bg-green-500 hover:bg-green-600 text-white font-medium px-6 py-3 rounded-lg text-sm transition-colors mb-4 flex items-center justify-center gap-2">
          {{ 'schedule.share' | translate }}
        </button>

        <!-- Date tabs -->
        <div class="flex gap-2 flex-wrap mb-4" *ngIf="dates.length > 1">
          <button *ngFor="let d of dates" (click)="selectedDate=d"
            class="px-4 py-2 rounded-full text-sm font-medium transition-colors"
            [class]="selectedDate===d
              ? 'bg-indigo-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'">
            {{ d }}
          </button>
        </div>

        <!-- No slots -->
        <p *ngIf="slotsForDate.length===0"
           class="text-gray-400 text-center py-12 text-sm">
          {{ 'schedule.noSlots' | translate }}
        </p>

        <!-- Mobile: grouped by stage -->
        <div *ngIf="slotsForDate.length>0" class="block lg:hidden space-y-4">
          <div *ngFor="let stage of stagesForDate"
               class="bg-white border border-gray-200 rounded-lg overflow-hidden shadow-sm">
            <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-100 bg-gray-50">
              <span class="w-3 h-3 rounded-full shrink-0" [style.background]="stageColor(stage.id)"></span>
              <span class="text-sm font-semibold text-gray-800">{{ stage.name }}</span>
            </div>
            <div class="divide-y divide-gray-100">
              <div *ngFor="let slot of slotsForStage(stage.id)"
                   class="flex items-center gap-3 px-4 py-3">
                <span class="text-xs text-gray-400 w-24 shrink-0">
                  {{ slot.start_time }}–{{ slot.end_time }}
                </span>
                <span class="text-sm font-medium text-gray-900">{{ slot.dj_name || '—' }}</span>
                <span *ngIf="slot.notes" class="text-xs text-gray-400 ml-auto">{{ slot.notes }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Desktop: time grid -->
        <div *ngIf="slotsForDate.length>0" class="hidden lg:block overflow-x-auto">
          <table class="w-full border-collapse text-sm">
            <thead>
              <tr>
                <th class="border border-gray-200 bg-gray-50 px-3 py-2 text-left w-36 text-xs font-medium text-gray-600">
                  {{ 'slots.stage' | translate }}
                </th>
                <th *ngFor="let t of timeHeaders"
                    class="border border-gray-200 bg-gray-50 px-2 py-2 text-xs font-medium text-gray-500 w-20 text-center">
                  {{ t }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let stage of stagesForDate">
                <td class="border border-gray-200 bg-gray-50 px-3 py-2">
                  <div class="flex items-center gap-2">
                    <span class="w-2.5 h-2.5 rounded-full shrink-0" [style.background]="stageColor(stage.id)"></span>
                    <span class="text-xs font-medium text-gray-800">{{ stage.name }}</span>
                  </div>
                </td>
                <td *ngFor="let t of timeHeaders"
                    class="border border-gray-200 px-1 py-1 min-w-[5rem] align-top">
                  <ng-container *ngFor="let slot of slotsAt(stage.id, t)">
                    <div class="rounded px-2 py-1 text-white text-xs font-medium leading-tight"
                         [style.background]="stageColor(stage.id)">
                      {{ slot.dj_name || '—' }}
                    </div>
                  </ng-container>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,
})
export class ScheduleComponent implements OnInit {
  event: Event | null = null;
  stages: Stage[] = [];
  slots: Slot[] = [];
  dates: string[] = [];
  selectedDate = '';
  timeHeaders: string[] = [];
  currentLang = 'en';

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private translate: TranslateService,
  ) {}

  ngOnInit() {
    const saved = localStorage.getItem('lang') || 'en';
    this.translate.use(saved);
    this.currentLang = saved;

    const eventId = this.route.snapshot.paramMap.get('id')!;
    this.api.getEvent(eventId).subscribe(e => {
      this.event = e;
      this.dates = this.dateRange(e.start_date, e.end_date);
      if (this.dates.length > 0) this.selectedDate = this.dates[0];
    });
    this.api.getStages(eventId).subscribe(s => this.stages = s);
    this.api.getSlots(eventId).subscribe(s => {
      this.slots = s;
      this.buildTimeHeaders();
    });
  }

  get slotsForDate(): Slot[] {
    return this.slots.filter(s => s.slot_date === this.selectedDate);
  }

  get stagesForDate(): Stage[] {
    const ids = new Set(this.slotsForDate.map(s => s.stage_id));
    return this.stages.filter(s => ids.has(s.id));
  }

  slotsForStage(stageId: string): Slot[] {
    return this.slotsForDate
      .filter(s => s.stage_id === stageId)
      .sort((a, b) => a.start_time.localeCompare(b.start_time));
  }

  slotsAt(stageId: string, time: string): Slot[] {
    return this.slotsForDate.filter(s =>
      s.stage_id === stageId && s.start_time <= time && s.end_time > time);
  }

  stageColor(stageId: string): string {
    return this.stages.find(s => s.id === stageId)?.color ?? '#6366F1';
  }

  buildTimeHeaders() {
    if (this.slots.length === 0) {
      this.timeHeaders = this.generateHeaders('14:00', '23:00');
      return;
    }
    const starts = this.slots.map(s => s.start_time).sort();
    const ends = this.slots.map(s => s.end_time).sort();
    const earliest = starts[0];
    const latest = ends[ends.length - 1];
    // Start 30 min before first slot, end at last slot end
    const startMin = this.toMinutes(earliest) - 30;
    const endMin = this.toMinutes(latest);
    this.timeHeaders = this.generateHeadersFromMinutes(
      Math.max(startMin, 0), endMin);
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
    // Round down to nearest 30
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
    this.translate.use(lang);
    this.currentLang = lang;
    localStorage.setItem('lang', lang);
  }

  shareViaLine() {
    const url = encodeURIComponent(window.location.href);
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }
}
