import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { provideTranslateService } from '@ngx-translate/core';

import { ScheduleComponent } from './schedule.component';
import { ApiService, PublicSchedule } from '../services/api.service';
import { LanguageService } from '../services/language.service';

const SCHEDULE: PublicSchedule = {
  event: {
    id: 'evt-1',
    name: 'Spring Showcase',
    venue_name: 'Revolver',
    start_date: '2026-07-12',
    end_date: '2026-07-12',
    genres: [],
  },
  stages: [
    { id: 'st-1', event_id: 'evt-1', name: 'Main Stage', color: '#ff9e2c', display_order: 0 },
    { id: 'st-2', event_id: 'evt-1', name: 'Booth 2', color: '#22d3ee', display_order: 1 },
  ],
  slots: [
    {
      id: 'slot-1', event_id: 'evt-1', stage_id: 'st-1', stage_name: 'Main Stage',
      dj_id: 'dj-1', dj_name: 'Cody', genre: 'House',
      slot_date: '2026-07-12', start_time: '22:00', end_time: '23:00',
      notes: '', dj_confirmation: null,
    },
  ],
};

describe('ScheduleComponent (EL-011 / EL-082 reskin)', () => {
  let getPublicSchedule: ReturnType<typeof vi.fn>;

  async function build(): Promise<{ fixture: ComponentFixture<ScheduleComponent>; component: ScheduleComponent }> {
    getPublicSchedule = vi.fn().mockReturnValue(of(SCHEDULE));
    const apiMock = { getPublicSchedule };

    await TestBed.configureTestingModule({
      imports: [ScheduleComponent],
      providers: [
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: LanguageService, useValue: { setLanguage: () => {} } },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ScheduleComponent);
    const component = fixture.componentInstance;
    fixture.detectChanges();
    return { fixture, component };
  }

  // The public schedule is unauthenticated: it MUST read from the public endpoint,
  // never a protected one (the EL-034 regression: protected endpoints 401 a
  // visitor and bounce them to /login). The reskin must not change this.
  it('reads the schedule from the public (unauthenticated) API', async () => {
    const { component } = await build();
    expect(getPublicSchedule).toHaveBeenCalledTimes(1);
    expect(getPublicSchedule).toHaveBeenCalledWith('evt-1');
    expect(component.event()?.name).toBe('Spring Showcase');
    expect(component.stages().length).toBe(2);
    expect(component.slots().length).toBe(1);
  });

  it('selects the first date and exposes only stages that have slots that day', async () => {
    const { component } = await build();
    expect(component.selectedDate()).toBe('2026-07-12');
    // Only Main Stage has a slot on the selected date; Booth 2 is filtered out.
    expect(component.stagesForDate().map(s => s.id)).toEqual(['st-1']);
  });

  it('resolves a stage colour, falling back for an unknown stage', async () => {
    const { component } = await build();
    expect(component.stageColor('st-1')).toBe('#ff9e2c');
    expect(component.stageColor('nope')).toBe('#6366F1');
  });
});
