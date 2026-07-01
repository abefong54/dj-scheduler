import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { PerformanceComponent } from './performance.component';
import { ApiService, RosterPerformance, DJPerformance } from '../../services/api.service';

const ROSTER: RosterPerformance[] = [
  { dj_id: 'a', dj_name: 'Ana', is_student: true, reps: 0, total_minutes: 0, last_played: '' },
  { dj_id: 'b', dj_name: 'Ben', is_student: true, reps: 5, total_minutes: 150, last_played: '2026-07-03' },
];
const UNDER: RosterPerformance[] = [ROSTER[0]];
const PERF: DJPerformance = {
  dj_id: 'b', dj_name: 'Ben', reps: 5, total_minutes: 150, last_played: '2026-07-03',
  by_genre: [
    { genre: 'House', reps: 3, total_minutes: 90 },
    { genre: 'Techno', reps: 2, total_minutes: 60 },
  ],
};

function makeApi() {
  return {
    getPerformance: vi.fn(() => of(ROSTER)),
    getUnderserved: vi.fn(() => of(UNDER)),
    getDJPerformance: vi.fn(() => of(PERF)),
  };
}

function render(api = makeApi()) {
  TestBed.configureTestingModule({
    imports: [PerformanceComponent],
    providers: [
      provideRouter([]),
      provideTranslateService(),
      { provide: ApiService, useValue: api },
    ],
  });
  const fixture = TestBed.createComponent(PerformanceComponent);
  fixture.detectChanges();
  return fixture;
}

describe('PerformanceComponent — Console shell', () => {
  it('wraps the page in the AdminShell with the performance nav active', () => {
    const root = render().nativeElement as HTMLElement;
    expect(root.querySelector('app-admin-shell')).toBeTruthy();
    expect(root.querySelector('.shell-sidebar')).toBeTruthy();
    expect(
      root.querySelector('[data-nav="performance"]')?.classList.contains('shell-nav-active'),
    ).toBe(true);
  });
});

describe('PerformanceComponent — roster + under-served (EL-044)', () => {
  it('keeps zero-rep students in the roster and flags under-served ones', () => {
    const c = render().componentInstance;
    const rows = c.rows();
    expect(rows.map(r => r['id'] as string)).toEqual(['a', 'b']); // zero-rep "a" not dropped
    expect(rows.find(r => r['id'] === 'a')!['underserved']).toBe(true);
    const ben = rows.find(r => r['id'] === 'b')!;
    expect(ben['underserved']).toBe(false);
    expect(ben['stageTime']).toBe('2h 30m');
  });

  it('sums total reps across the roster', () => {
    expect(render().componentInstance.totalReps()).toBe(5);
  });

  it('renders the under-served callout with the flagged student', () => {
    const root = render().nativeElement as HTMLElement;
    expect(root.querySelector('[data-testid="underserved-a"]')).toBeTruthy();
  });

  // EL-081: the fairness signal sits on token-driven Soundcheck surfaces — the
  // hard-coded light-amber wrapper is gone, replaced by the reskin classes.
  it('renders the under-served signal on token-driven surfaces (no hard-coded light amber)', () => {
    const root = render().nativeElement as HTMLElement;
    const row = root.querySelector('[data-testid="underserved-a"]') as HTMLElement;
    expect(row).toBeTruthy();
    expect(row.classList).toContain('underserved-row');
    expect(row.classList.contains('bg-amber-50')).toBe(false);
    expect(row.querySelector('.underserved-dot')).toBeTruthy();
  });
});

describe('PerformanceComponent — drill-in', () => {
  it('viewDj loads the DJ via the per-DJ endpoint; closeDrill clears it', () => {
    const api = makeApi();
    const c = render(api).componentInstance;
    c.viewDj('b');
    expect(api.getDJPerformance).toHaveBeenCalledWith('b');
    expect(c.selectedPerf()?.by_genre.length).toBe(2);
    c.closeDrill();
    expect(c.selectedPerf()).toBeNull();
  });
});

describe('PerformanceComponent — fmtMinutes', () => {
  it('formats minutes as h/m', () => {
    const c = render().componentInstance;
    expect(c.fmtMinutes(0)).toBe('0m');
    expect(c.fmtMinutes(45)).toBe('45m');
    expect(c.fmtMinutes(60)).toBe('1h');
    expect(c.fmtMinutes(150)).toBe('2h 30m');
  });
});
