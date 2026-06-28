import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';
import { provideTranslateService, TranslateService } from '@ngx-translate/core';

import { EventDetailComponent } from './event-detail.component';
import { ApiService, Slot } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';
import { ScheduleExportService } from '../../services/schedule-export.service';

const EVENT = {
  id: 'evt-1',
  name: 'Test Event',
  venue_name: 'Venue',
  start_date: '2026-07-01',
  end_date: '2026-07-01',
  genres: [],
};
const STAGES = [
  { id: 'st-1', event_id: 'evt-1', name: 'Main Stage', color: '#000', display_order: 0 },
];
const DJS = [{ id: 'dj-1', name: 'DJ Alpha', genre_tags: ['House'] }];

const SLOT: Slot = {
  id: 'slot-1',
  event_id: 'evt-1',
  stage_id: 'st-1',
  stage_name: 'Main Stage',
  dj_id: 'dj-1',
  dj_name: 'DJ Alpha',
  genre: 'House',
  slot_date: '2026-07-01',
  start_time: '19:00',
  end_time: '20:00',
  notes: '',
} as Slot;

// Builds an Angular-style HttpErrorResponse-like object.
const httpError = (status: number, body: unknown) => ({ status, error: body });

describe('EventDetailComponent — conflict warnings (US-005)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  let component: EventDetailComponent;

  // Per-test swappable responses for the slot write endpoints.
  let createSlot$: any;
  let updateSlot$: any;
  let slotsResponse: Slot[];

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of(slotsResponse),
    getDJs: () => of(DJS),
    createSlot: () => createSlot$,
    updateSlot: () => updateSlot$,
  };

  beforeEach(async () => {
    slotsResponse = [];
    createSlot$ = of({});
    updateSlot$ = of({});

    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();

    const translate = TestBed.inject(TranslateService);
    translate.setTranslation('en', {
      'slots.error.djDoubleBooked': '{{djName}} is already booked at this time',
      'slots.error.stageOverlap': '{{stageName}} already has a set at this time',
      'slots.error.generic': 'Something went wrong, please try again',
    });
    translate.use('en');

    fixture = TestBed.createComponent(EventDetailComponent);
    component = fixture.componentInstance;
  });

  function fillAddRow() {
    component.activateAddRow();
    component.addStageId = 'st-1';
    component.addDjId = 'dj-1';
    component.addGenre = 'House';
    component.addDate = '2026-07-01';
    component.addStart = '19:00';
    component.addDuration = 60;
  }

  it('shows the DJ name on a dj_double_booked conflict (409)', () => {
    fillAddRow();
    createSlot$ = throwError(() => httpError(409, { type: 'dj_double_booked', conflicting_slot_id: 'x' }));
    component.submitAddRow();
    expect(component.addConflictError()).toBe('DJ Alpha is already booked at this time');
  });

  it('shows the stage name on a stage_overlap conflict (409)', () => {
    fillAddRow();
    createSlot$ = throwError(() => httpError(409, { type: 'stage_overlap', conflicting_slot_id: 'x' }));
    component.submitAddRow();
    expect(component.addConflictError()).toBe('Main Stage already has a set at this time');
  });

  it('shows a generic error on a 500', () => {
    fillAddRow();
    createSlot$ = throwError(() => httpError(500, {}));
    component.submitAddRow();
    expect(component.addConflictError()).toBe('Something went wrong, please try again');
  });

  it('preserves the form inputs when a conflict occurs', () => {
    fillAddRow();
    createSlot$ = throwError(() => httpError(409, { type: 'dj_double_booked' }));
    component.submitAddRow();
    expect(component.addStart).toBe('19:00');
    expect(component.addDjId).toBe('dj-1');
    expect(component.addStageId).toBe('st-1');
    expect(component.addRowActive()).toBe(true);
  });

  it('clears the error after a successful resubmission', () => {
    fillAddRow();
    createSlot$ = throwError(() => httpError(409, { type: 'dj_double_booked' }));
    component.submitAddRow();
    expect(component.addConflictError()).not.toBeNull();

    // Organizer fixes the time and resubmits successfully.
    createSlot$ = of({ ...SLOT, id: 'slot-2', start_time: '21:00' });
    component.addStart = '21:00';
    component.submitAddRow();
    expect(component.addConflictError()).toBeNull();
  });

  it('surfaces conflicts on inline edit save as well', () => {
    component.startEdit(SLOT);
    updateSlot$ = throwError(() => httpError(409, { type: 'stage_overlap' }));
    component.saveEdit('slot-1');
    expect(component.editConflictError()).toBe('Main Stage already has a set at this time');
  });

  it('clears the edit error on cancel', () => {
    component.startEdit(SLOT);
    updateSlot$ = throwError(() => httpError(409, { type: 'stage_overlap' }));
    component.saveEdit('slot-1');
    expect(component.editConflictError()).not.toBeNull();
    component.cancelEdit();
    expect(component.editConflictError()).toBeNull();
  });
});
