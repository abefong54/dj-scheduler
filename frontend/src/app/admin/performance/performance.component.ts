import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, RosterPerformance, DJPerformance } from '../../services/api.service';
import { LanguageService } from '../../services/language.service';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';
import { ButtonComponent } from '../../shared/button.component';
import { AdminShellComponent } from '../../shared/admin-shell.component';

/**
 * Performance dashboard (EL-044, Wedge B). The teacher-facing view of who is
 * getting stage time: an under-served callout (students below the rep threshold,
 * surfaced first), a roster table that keeps zero-rep students visible, and a
 * per-DJ drill-in with a by-genre breakdown. Consumes EL-043's three endpoints.
 */
@Component({
  selector: 'app-performance',
  standalone: true,
  imports: [TranslatePipe, DataTableComponent, ColumnDefDirective, ButtonComponent, AdminShellComponent],
  templateUrl: './performance.component.html',
})
export class PerformanceComponent implements OnDestroy {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private langService = inject(LanguageService);

  // GET /api/performance returns the roster already ordered reps-ascending, so
  // source order surfaces under-served students first by default.
  roster = signal<RosterPerformance[]>([]);
  underserved = signal<RosterPerformance[]>([]);
  selectedPerf = signal<DJPerformance | null>(null);

  totalReps = computed(() => this.roster().reduce((sum, r) => sum + r.reps, 0));

  private subscriptions: Subscription[] = [];

  constructor() {
    this.load();
  }

  private load() {
    this.subscriptions.push(
      this.api.getPerformance().subscribe(r => this.roster.set(r)),
      this.api.getUnderserved().subscribe(u => this.underserved.set(u)),
    );
  }

  // Rows for DataTableComponent: needs a unique `id` and plain record fields.
  // `underserved` flags rows for the visual marker; stage time is pre-formatted.
  rows = computed((): Record<string, unknown>[] => {
    const under = new Set(this.underserved().map(u => u.dj_id));
    return this.roster().map(r => ({
      id: r.dj_id,
      name: r.dj_name,
      reps: r.reps,
      stageTime: this.fmtMinutes(r.total_minutes),
      lastPlayed: r.last_played,
      underserved: under.has(r.dj_id),
    }) as Record<string, unknown>);
  });

  // Reactive labels re-evaluate on language switch. reps/stageTime are numeric so
  // they're left non-sortable (the table sorts lexically); the default reps-asc
  // source order is what matters here anyway.
  columns = computed((): TableColumn[] => {
    this.langService.currentLang();
    return [
      { key: 'name', label: this.translate.instant('performance.col.student') },
      { key: 'reps', label: this.translate.instant('performance.col.reps'), sortable: false, searchable: false },
      { key: 'stageTime', label: this.translate.instant('performance.col.stageTime'), sortable: false, searchable: false },
      { key: 'lastPlayed', label: this.translate.instant('performance.col.lastPlayed'), searchable: false },
      { key: 'actions', label: this.translate.instant('performance.col.actions'), sortable: false, searchable: false },
    ];
  });

  rowId(row: Record<string, unknown>): string {
    return String(row['id']);
  }

  viewDj(id: string) {
    this.subscriptions.push(
      this.api.getDJPerformance(id).subscribe(p => this.selectedPerf.set(p)),
    );
  }

  closeDrill() {
    this.selectedPerf.set(null);
  }

  // Minutes → "Xh Ym" / "Xh" / "Ym". 0 → "0m".
  fmtMinutes(total: number): string {
    const m = Math.max(0, Math.round(total ?? 0));
    const h = Math.floor(m / 60);
    const min = m % 60;
    if (h && min) return `${h}h ${min}m`;
    if (h) return `${h}h`;
    return `${min}m`;
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
