import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { provideTranslateService, TranslateService } from '@ngx-translate/core';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { EventDetailComponent } from './event-detail.component';
import { ApiService, Slot } from '../../services/api.service';
import { EventsStore } from '../../services/events-store';
import { DialogService } from '../../shared/dialog.service';
import { ScheduleExportService } from '../../services/schedule-export.service';
import { BreakpointService } from '../../shared/breakpoint.service';

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

// EL-042: certification-gated slot assignment.
describe('EventDetailComponent — certification gating (EL-042)', () => {
  const MIA = { id: 'mia', name: 'Mia', genre_tags: ['Hip Hop'], certifications: [], is_student: true };
  const KAI = { id: 'kai', name: 'Kai', genre_tags: ['Hip Hop'], certifications: ['Hip Hop'], is_student: true };
  const PRO = { id: 'pro', name: 'Pro', genre_tags: ['Hip Hop'], certifications: [], is_student: false };

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of([] as Slot[]),
    getDJs: () => of([MIA, KAI, PRO]),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  let component: EventDetailComponent;

  beforeEach(async () => {
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
      'slots.cert.warning': "{{djName}} isn't certified for {{genre}}",
    });
    translate.use('en');

    component = TestBed.createComponent(EventDetailComponent).componentInstance;
  });

  it('treats a student without the certification as uncertified', () => {
    expect(component.isCertifiedFor(MIA, 'Hip Hop')).toBe(false);
    expect(component.isCertifiedFor(KAI, 'Hip Hop')).toBe(true);
  });

  it('lets graduates bypass the gate', () => {
    expect(component.isCertifiedFor(PRO, 'Hip Hop')).toBe(true);
  });

  it('matches certifications case-insensitively', () => {
    expect(component.isCertifiedFor(KAI, 'hip hop')).toBe(true);
  });

  it('orders certified DJs first and flags the rest (not hidden)', () => {
    component.addGenre = 'Hip Hop';
    const opts = component.djOptionsForAdd();
    // All three remain selectable (uncertified are flagged, not hidden).
    expect(opts.map(o => o.dj.id).sort()).toEqual(['kai', 'mia', 'pro']);
    // The uncertified student sorts last.
    expect(opts[opts.length - 1].dj.id).toBe('mia');
    expect(opts.find(o => o.dj.id === 'mia')!.certified).toBe(false);
  });

  it('warns for an uncertified student, but not for certified DJs or graduates', () => {
    component.addGenre = 'Hip Hop';

    component.addDjId = 'mia';
    expect(component.addCertWarning()).toContain('Mia');

    component.addDjId = 'kai';
    expect(component.addCertWarning()).toBeNull();

    component.addDjId = 'pro';
    expect(component.addCertWarning()).toBeNull();
  });
});

// EL-066: Console restyle — the page renders inside the shared AdminShell with
// a monumental event name, mono meta readout, mono table headers, and a Console
// status badge. Presentation only; the behavior suites above guard the logic.
describe('EventDetailComponent — Console restyle (EL-066)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  const root = () => fixture.nativeElement as HTMLElement;

  const CONFIRMED_SLOT: Slot = { ...SLOT, dj_confirmation: 'confirmed' } as Slot;

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of([CONFIRMED_SLOT]),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();

    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    fixture.detectChanges();
  });

  it('renders inside the AdminShell with the events nav active', () => {
    expect(root().querySelector('app-admin-shell')).toBeTruthy();
    expect(root().querySelector('.shell-nav-active')?.getAttribute('data-nav')).toBe('events');
  });

  it('shows the event name as a monument and a mono meta readout', () => {
    expect(root().querySelector('.detail-event-name')?.textContent).toContain('Test Event');
    expect(root().querySelector('.detail-event-meta')?.textContent).toContain('Venue');
  });

  it('renders mono uppercase column headers and a Console status badge', () => {
    expect(root().querySelectorAll('.slot-th').length).toBeGreaterThan(0);
    expect(root().querySelector('app-status-badge .status-badge.status-confirmed')).toBeTruthy();
  });

  it('keeps the inline add-slot trigger wired', () => {
    expect(root().querySelector('[data-testid="add-slot-trigger"]')).toBeTruthy();
  });
});

// EL-080: Soundcheck reskin — the page body renders the token-driven surface
// classes it overrides onto the dark-booth palette. Skin only; these guard that
// the reskinned hooks (slots heading, dark modal) stay present so the component
// CSS overrides keep landing. Supersedes the EL-066 Console restyle.
describe('EventDetailComponent — Soundcheck reskin (EL-080)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  const root = () => fixture.nativeElement as HTMLElement;

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of([SLOT]),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();

    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    fixture.detectChanges();
  });

  it('renders the reskinned slots section heading hook', () => {
    expect(root().querySelector('.slots-section .section-title')).toBeTruthy();
  });

  it('renders the add-stage modal on the dark card surface with swatches', () => {
    fixture.componentInstance.addStage();
    fixture.detectChanges();
    expect(root().querySelector('.modal-box-sm')).toBeTruthy();
    expect(root().querySelector('.modal-box-sm .modal-title')).toBeTruthy();
    expect(root().querySelectorAll('.color-swatch').length).toBeGreaterThan(0);
  });
});

// EL-028: slots table search + sort via the shared TableSort engine. The merged
// template renders inside the AdminShell (RouterLink), so this suite provides the
// router too even though it only asserts on the engine.
describe('EventDetailComponent — slot search & sort (EL-028)', () => {
  const slots: Slot[] = [
    { ...SLOT, id: 's-a', dj_name: 'DJ Alpha', stage_name: 'Main Stage', genre: 'House', start_time: '21:00' } as Slot,
    { ...SLOT, id: 's-b', dj_name: 'DJ Beta', stage_name: 'Side Stage', genre: 'Techno', start_time: '19:00' } as Slot,
    { ...SLOT, id: 's-c', dj_name: 'MC Gamma', stage_name: 'Main Stage', genre: 'House', start_time: '20:00' } as Slot,
  ];

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of(slots),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  let component: EventDetailComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();

    component = TestBed.createComponent(EventDetailComponent).componentInstance;
  });

  it('defaults to ascending start_time order', () => {
    expect(component.slotTable.view().map(s => s.id)).toEqual(['s-b', 's-c', 's-a']);
    expect(component.slotTable.indicator('start_time')).toBe('↑');
  });

  it('filters across dj, stage, and genre case-insensitively', () => {
    component.slotTable.search.set('side');
    expect(component.slotTable.view().map(s => s.id)).toEqual(['s-b']);

    component.slotTable.search.set('techno');
    expect(component.slotTable.view().map(s => s.id)).toEqual(['s-b']);
  });

  it('toggles a column to sort and reverses on a second click', () => {
    component.slotTable.toggle('dj_name');
    expect(component.slotTable.view().map(s => s.dj_name)).toEqual(['DJ Alpha', 'DJ Beta', 'MC Gamma']);
    component.slotTable.toggle('dj_name');
    expect(component.slotTable.view().map(s => s.dj_name)).toEqual(['MC Gamma', 'DJ Beta', 'DJ Alpha']);
    expect(component.slotTable.indicator('dj_name')).toBe('↓');
  });

  it('sorts only the filtered rows when search and sort compose', () => {
    component.slotTable.search.set('dj '); // DJ Alpha + DJ Beta
    component.slotTable.toggle('dj_name');
    component.slotTable.sortDir.set('desc');
    expect(component.slotTable.view().map(s => s.dj_name)).toEqual(['DJ Beta', 'DJ Alpha']);
  });
});

// Hero editor tests (EL-092): Edit button on the detail hero flips name/venue/
// dates/genres into inputs; Save PATCHes via updateEvent, Cancel discards.
// Uses HttpClientTestingModule so we can assert the exact PATCH payload.
describe('EventDetailComponent — hero editor', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true), alert: () => {} } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();

    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    httpMock = TestBed.inject(HttpTestingController);

    // Flush the 4 initialization requests so event() is populated.
    httpMock.expectOne('/api/events/evt-1').flush(EVENT);
    httpMock.expectOne('/api/events/evt-1/stages').flush(STAGES);
    httpMock.expectOne('/api/events/evt-1/slots').flush([]);
    httpMock.expectOne('/api/djs').flush(DJS);
  });

  afterEach(() => httpMock.verify());

  it('opens the hero editor seeded from the event', () => {
    const cmp = fixture.componentInstance;
    cmp.startHeaderEdit();
    expect(cmp.editingHeader()).toBe(true);
    expect(cmp.hVenue()).toBe(cmp.event()!.venue_name);
  });

  it('saveHeader PATCHes the event and closes the editor', () => {
    const store = TestBed.inject(EventsStore);
    vi.spyOn(store, 'refresh').mockImplementation(() => {}); // prevent the implicit getEvents() from leaking into verify()
    const cmp = fixture.componentInstance;
    cmp.startHeaderEdit();
    cmp.hVenue.set('Renamed Venue');
    cmp.saveHeader();
    const req = httpMock.expectOne(`/api/events/${cmp.event()!.id}`);
    expect(req.request.method).toBe('PATCH');
    expect(req.request.body.venue_name).toBe('Renamed Venue');
    req.flush({ ...req.request.body, id: cmp.event()!.id });
    fixture.detectChanges();
    expect(cmp.editingHeader()).toBe(false);
    expect(cmp.event()!.venue_name).toBe('Renamed Venue');
  });

  it('saveHeader calls EventsStore.refresh() on success', () => {
    const store = TestBed.inject(EventsStore);
    const refreshSpy = vi.spyOn(store, 'refresh').mockImplementation(() => {});
    const cmp = fixture.componentInstance;
    cmp.startHeaderEdit();
    cmp.hVenue.set('New Venue');
    cmp.saveHeader();
    const req = httpMock.expectOne(`/api/events/${cmp.event()!.id}`);
    req.flush({ ...req.request.body, id: cmp.event()!.id });
    expect(refreshSpy).toHaveBeenCalled();
  });

  it('cancelHeader closes without a request', () => {
    const cmp = fixture.componentInstance;
    cmp.startHeaderEdit();
    cmp.hVenue.set('discard me');
    cmp.cancelHeader();
    httpMock.expectNone(`/api/events/${cmp.event()!.id}`);
    expect(cmp.editingHeader()).toBe(false);
  });
});

// EL-094: below md the table becomes stacked cards and add/edit moves into a
// dialog bound to the same state/methods. A stub BreakpointService drives the
// compact() switch.
describe('EventDetailComponent — mobile cards + dialog (EL-094)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  let compact: WritableSignal<boolean>;
  const root = () => fixture.nativeElement as HTMLElement;

  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of([SLOT]),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  beforeEach(async () => {
    compact = signal(true);
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
        { provide: BreakpointService, useValue: { compact } },
      ],
    }).compileComponents();

    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    fixture.detectChanges();
  });

  it('renders stacked slot cards (not the table) when compact', () => {
    expect(root().querySelectorAll('.slot-card').length).toBe(1);
    expect(root().querySelector('.slot-table')).toBeNull();
  });

  it('renders the table (not cards) when not compact', () => {
    compact.set(false);
    fixture.detectChanges();
    expect(root().querySelector('.slot-table')).toBeTruthy();
    expect(root().querySelector('.slot-card')).toBeNull();
  });

  it('opens the add-slot dialog from the card-list trigger', () => {
    root().querySelector<HTMLButtonElement>('[data-testid="add-slot-trigger"]')!.click();
    fixture.detectChanges();
    expect(root().querySelector('[data-testid="add-slot-dialog"]')).toBeTruthy();
  });

  it('edits a slot via the dialog through the shared saveEdit path', () => {
    fixture.componentInstance.startEdit(SLOT);
    fixture.detectChanges();
    expect(root().querySelector('[data-testid="edit-slot-dialog"]')).toBeTruthy();

    // Saving through the dialog runs the same saveEdit → updateSlot path; on
    // success the dialog closes.
    root().querySelector<HTMLButtonElement>('[data-testid="dialog-edit-save"]')!.click();
    fixture.detectChanges();
    expect(fixture.componentInstance.editingSlotId()).toBeNull();
    expect(root().querySelector('[data-testid="edit-slot-dialog"]')).toBeNull();
  });
});

// EL-095: interaction backfill — date tabs and the add-stage colour swatch had
// no coverage. (Sort headers, keyboard save/cancel, and swatch→create end-to-end
// are covered by e2e/specs/planner-interactions.spec.ts.)
describe('EventDetailComponent — date tabs (EL-095)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  const root = () => fixture.nativeElement as HTMLElement;
  // Two slots on two different dates → the date tabs render.
  const MULTI: Slot[] = [
    { ...SLOT, id: 'd1', slot_date: '2026-07-01', start_time: '19:00', end_time: '20:00' } as Slot,
    { ...SLOT, id: 'd2', slot_date: '2026-07-02', start_time: '20:00', end_time: '21:00' } as Slot,
  ];
  const apiMock = {
    getEvent: () => of({ ...EVENT, end_date: '2026-07-02' }),
    getStages: () => of(STAGES),
    getSlots: () => of(MULTI),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();
    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    fixture.detectChanges();
  });

  it('renders one tab per date and defaults to the first', () => {
    expect(root().querySelectorAll('.tab-btn').length).toBe(2);
    expect(fixture.componentInstance.selectedDate()).toBe('2026-07-01');
    expect(fixture.componentInstance.slotTable.view().map(s => s.id)).toEqual(['d1']);
  });

  it('switches the visible date when another tab is clicked', () => {
    root().querySelectorAll<HTMLButtonElement>('.tab-btn')[1].click();
    fixture.detectChanges();
    expect(fixture.componentInstance.selectedDate()).toBe('2026-07-02');
    expect(fixture.componentInstance.slotTable.view().map(s => s.id)).toEqual(['d2']);
  });
});

describe('EventDetailComponent — add-stage colour swatch (EL-095)', () => {
  let fixture: ComponentFixture<EventDetailComponent>;
  const root = () => fixture.nativeElement as HTMLElement;
  const createStage = vi.fn(() => of({}));
  const apiMock = {
    getEvent: () => of(EVENT),
    getStages: () => of(STAGES),
    getSlots: () => of([] as Slot[]),
    getDJs: () => of(DJS),
    createSlot: () => of({}),
    updateSlot: () => of({}),
    createStage,
  };

  beforeEach(async () => {
    createStage.mockClear();
    await TestBed.configureTestingModule({
      imports: [EventDetailComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'evt-1' } } } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
        { provide: ScheduleExportService, useValue: { download: () => {} } },
      ],
    }).compileComponents();
    TestBed.inject(TranslateService).use('en');
    fixture = TestBed.createComponent(EventDetailComponent);
    fixture.detectChanges();
  });

  it('marks the chosen swatch active and creates the stage with that colour', () => {
    const comp = fixture.componentInstance;
    comp.addStage();
    fixture.detectChanges();

    const swatches = root().querySelectorAll<HTMLButtonElement>('.color-swatch');
    expect(swatches.length).toBe(comp.stageColors.length);
    swatches[3].click();
    fixture.detectChanges();

    expect(comp.newStageColor).toBe(comp.stageColors[3]);
    expect(swatches[3].className).toContain('swatch-active');

    comp.newStageName = 'Night Stage';
    root().querySelector<HTMLButtonElement>('[data-testid="stage-create-btn"]')!.click();
    expect(createStage).toHaveBeenCalledWith('evt-1', { name: 'Night Stage', color: comp.stageColors[3] });
  });
});
