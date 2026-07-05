import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of, Subject, throwError } from 'rxjs';
import { provideTranslateService } from '@ngx-translate/core';
import { vi } from 'vitest';

import { EventsListComponent } from './events-list.component';
import { ApiService, Event } from '../../services/api.service';
import { EventsStore } from '../../services/events-store';
import { DialogService } from '../../shared/dialog.service';

// Deterministic "today" anchor; the component's `today` field is overridden below.
const TODAY = '2026-06-30';

const LIVE_EVENT: Event = {
  id: 'evt-live',
  name: 'Pulse Festival',
  venue_name: 'Warehouse 9',
  start_date: '2026-06-29',
  end_date: '2026-07-01',
  genres: ['Techno', 'House'],
};
const FUTURE_EVENT: Event = {
  id: 'evt-future',
  name: 'Summer Sessions',
  venue_name: 'Rooftop Bar',
  start_date: '2026-08-01',
  end_date: '2026-08-01',
  genres: ['Disco'],
};
const PAST_EVENT: Event = {
  id: 'evt-past',
  name: 'Winter Rave',
  venue_name: 'The Vault',
  start_date: '2026-01-01',
  end_date: '2026-01-02',
  genres: [],
};

describe('EventsListComponent — Console restyle (EL-064)', () => {
  let fixture: ComponentFixture<EventsListComponent>;
  let component: EventsListComponent;
  const root = () => fixture.nativeElement as HTMLElement;

  const apiMock = {
    getEvents: () => of([LIVE_EVENT, FUTURE_EVENT, PAST_EVENT]),
    deleteEvent: () => of({}),
    cloneEvent: () => of({}),
    updateEvent: vi.fn(),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventsListComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve() },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(EventsListComponent);
    component = fixture.componentInstance;
    component.today = TODAY; // pin lifecycle math regardless of the real clock
    fixture.detectChanges();
  });

  it('wraps the page in the AdminShell with the Events nav active', () => {
    const shell = root().querySelector('app-admin-shell');
    expect(shell).toBeTruthy();
    expect(root().querySelector('.shell-sidebar')).toBeTruthy();
    expect(
      root().querySelector('[data-nav="events"]')?.classList.contains('shell-nav-active'),
    ).toBe(true);
  });

  it('renders stat tiles with real derived counts (upcoming / past / total)', () => {
    const numbers = Array.from(root().querySelectorAll('.stat-tile-number')).map((el) =>
      el.textContent?.trim(),
    );
    // 2 upcoming (live + future), 1 past, 3 total.
    expect(numbers).toEqual(['2', '1', '3']);
  });

  it('renders a header row with 6 columns via the shared data-table', () => {
    const headers = root().querySelectorAll('.dt-th');
    expect(headers.length).toBe(6);
  });

  it('renders one row per upcoming event (identified by its view link)', () => {
    expect(root().querySelector('[data-testid="event-view-evt-live"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-view-evt-future"]')).toBeTruthy();
    // Past event is on the Past tab, not the default Upcoming tab.
    expect(root().querySelector('[data-testid="event-view-evt-past"]')).toBeFalsy();
  });

  it('shows a LIVE status badge only for an in-progress event', () => {
    // Live event: status badge appears
    expect(root().querySelector('app-status-badge .status-live')).toBeTruthy();
    // Exactly one badge on the upcoming tab (only the live event, not the future one)
    expect(root().querySelectorAll('app-status-badge').length).toBe(1);
  });

  it('preserves navigation to the event detail page via the name link', () => {
    const link = root().querySelector(
      '[data-testid="event-view-evt-live"]',
    ) as HTMLAnchorElement | null;
    expect(link?.getAttribute('href')).toBe('/admin/events/evt-live');
  });

  it('switches to the Past tab and shows past events there', () => {
    component.activeTab.set('past');
    fixture.detectChanges();
    expect(root().querySelector('[data-testid="event-view-evt-past"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-view-evt-live"]')).toBeFalsy();
  });

  // EL-080: Soundcheck reskin — the tab strip carries the active/inactive hooks
  // the component CSS re-skins onto the dark-booth palette (amber active accent,
  // muted inactive). Skin only; this guards the markers stay present.
  it('renders the reskinned tab strip with active + inactive hooks', () => {
    expect(root().querySelector('.events-tabs')).toBeTruthy();
    expect(root().querySelector('.events-tab-active')).toBeTruthy();
    expect(root().querySelector('.events-tab-inactive')).toBeTruthy();
  });

  it('derives lifecycle purely from dates', () => {
    expect(component.lifecycle(LIVE_EVENT)).toBe('live');
    expect(component.lifecycle(FUTURE_EVENT)).toBe('upcoming');
    expect(component.lifecycle(PAST_EVENT)).toBe('past');
    expect(component.isLive(LIVE_EVENT)).toBe(true);
    expect(component.isLive(FUTURE_EVENT)).toBe(false);
  });

  it('renders the Actions column as icon-only buttons (edit, clone, delete)', () => {
    const editBtn = root().querySelector('[data-testid="event-edit-evt-live"]');
    const cloneBtn = root().querySelector('[data-testid="event-clone-evt-live"]');
    const deleteBtn = root().querySelector('[data-testid="event-delete-evt-live"]');

    // All three actions present, each rendered as an icon (no visible text label).
    for (const btn of [editBtn, cloneBtn, deleteBtn]) {
      expect(btn).toBeTruthy();
      expect(btn!.querySelector('app-icon')).toBeTruthy();
      expect(btn!.querySelector('span')).toBeFalsy(); // iconOnly hides the label span
    }

    // Accessible name preserved on the icon-only buttons: aria-label carries the
    // label key; the visible tooltip is now TooltipDirective (EL-090), not title.
    expect(editBtn!.getAttribute('aria-label')).toBe('actions.edit');
    expect(cloneBtn!.getAttribute('aria-label')).toBe('events.clone');
    expect(deleteBtn!.getAttribute('aria-label')).toBe('actions.delete');
  });

  it('renders the Actions column within the shared data-table', () => {
    // DataTableComponent renders all columns uniformly (dt-th / dt-td); action
    // buttons confirm the actions column is projected and rendered correctly.
    expect(root().querySelector('[data-testid="event-edit-evt-live"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-delete-evt-live"]')).toBeTruthy();
  });

  it('navigates to the event detail page when the edit icon is clicked', () => {
    const router = TestBed.inject(Router);
    const navSpy = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    const editBtn = root().querySelector(
      '[data-testid="event-edit-evt-live"]',
    ) as HTMLButtonElement;
    editBtn.click();

    expect(navSpy).toHaveBeenCalledWith(['/admin/events', 'evt-live']);
  });

  it('exposes edit() that routes to the detail page', () => {
    const router = TestBed.inject(Router);
    const navSpy = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    component.edit('evt-future');

    expect(navSpy).toHaveBeenCalledWith(['/admin/events', 'evt-future']);
  });

  it('shows edit + clone on every tab but delete only on Upcoming', () => {
    // Upcoming tab: all three actions on a live event.
    expect(root().querySelector('[data-testid="event-edit-evt-live"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-clone-evt-live"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-delete-evt-live"]')).toBeTruthy();

    component.activeTab.set('past');
    fixture.detectChanges();

    // Past tab: edit + clone remain, delete is hidden.
    expect(root().querySelector('[data-testid="event-edit-evt-past"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-clone-evt-past"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-delete-evt-past"]')).toBeFalsy();
  });
});

// EL-087: representative per-surface loading-state coverage. The gate pattern
// (loading signal + @if meter / @else if retry / @else content) is identical
// across every wired surface.
describe('EventsListComponent — loading state (EL-087)', () => {
  const setup = (getEvents: () => import('rxjs').Observable<Event[]>) => {
    TestBed.configureTestingModule({
      imports: [EventsListComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: { getEvents } },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve() },
        },
      ],
    });
    const fixture = TestBed.createComponent(EventsListComponent);
    fixture.detectChanges();
    return fixture;
  };

  it('shows the VU-meter while loading, then the content', () => {
    const subject = new Subject<Event[]>();
    const fixture = setup(() => subject.asObservable());
    const el = fixture.nativeElement as HTMLElement;

    expect(el.querySelector('app-loading-meter')).toBeTruthy();
    expect(el.querySelector('app-data-table')).toBeFalsy();

    // FUTURE_EVENT is upcoming vs. the real clock, so it lands on the default tab.
    subject.next([FUTURE_EVENT]);
    subject.complete();
    fixture.detectChanges();

    expect(el.querySelector('app-loading-meter')).toBeFalsy();
    expect(el.querySelector('app-data-table')).toBeTruthy();
  });

  it('shows a retry fallback when the load fails', () => {
    const subject = new Subject<Event[]>();
    const fixture = setup(() => subject.asObservable());
    const el = fixture.nativeElement as HTMLElement;

    subject.error(new Error('boom'));
    fixture.detectChanges();

    expect(el.querySelector('app-loading-meter')).toBeFalsy();
    expect(el.querySelector('button.btn-primary')?.textContent).toContain('common.retry');
  });
});

// EL-091: the list is backed by the root EventsStore (stale-while-revalidate), so
// a re-visit paints from cache with no spinner and every mutation refreshes it.
describe('EventsListComponent — cached list (EL-091)', () => {
  const configure = (api: Partial<ApiService>) =>
    TestBed.configureTestingModule({
      imports: [EventsListComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: api },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve() },
        },
      ],
    });

  it('paints cached rows with NO spinner on a second visit (revalidates silently)', () => {
    const first = new Subject<Event[]>();
    const second = new Subject<Event[]>(); // stays pending — cache must cover the gap
    let calls = 0;
    configure({
      getEvents: () => (++calls === 1 ? first : second).asObservable(),
    });

    // First visit: nothing cached → spinner, no table.
    const f1 = TestBed.createComponent(EventsListComponent);
    f1.detectChanges();
    const el1 = f1.nativeElement as HTMLElement;
    expect(el1.querySelector('app-loading-meter')).toBeTruthy();
    first.next([FUTURE_EVENT]);
    first.complete();
    f1.detectChanges();
    expect(el1.querySelector('app-data-table')).toBeTruthy();

    // Second visit shares the singleton store: cached rows show immediately even
    // though the revalidation request (`second`) is still in flight — no spinner.
    const f2 = TestBed.createComponent(EventsListComponent);
    f2.detectChanges();
    const el2 = f2.nativeElement as HTMLElement;
    expect(el2.querySelector('app-loading-meter')).toBeFalsy();
    expect(el2.querySelector('[data-testid="event-view-evt-future"]')).toBeTruthy();
  });

  it('refreshes the shared store after deleting an event', async () => {
    configure({ getEvents: () => of([FUTURE_EVENT]), deleteEvent: () => of({}) });
    const store = TestBed.inject(EventsStore);
    const refreshSpy = vi.spyOn(store, 'refresh');

    const fixture = TestBed.createComponent(EventsListComponent);
    fixture.detectChanges();
    await fixture.componentInstance.delete('evt-future');

    expect(refreshSpy).toHaveBeenCalled();
  });

  it('refreshes the shared store after cloning an event', async () => {
    configure({
      getEvents: () => of([FUTURE_EVENT]),
      cloneEvent: () => of({ ...FUTURE_EVENT, id: 'evt-clone' }),
    });
    const store = TestBed.inject(EventsStore);
    const refreshSpy = vi.spyOn(store, 'refresh');
    vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    const fixture = TestBed.createComponent(EventsListComponent);
    fixture.detectChanges();
    await fixture.componentInstance.clone('evt-future');

    expect(refreshSpy).toHaveBeenCalled();
  });

  it('refreshes the shared store after a successful inline edit (commitVenue)', () => {
    configure({
      getEvents: () => of([LIVE_EVENT, FUTURE_EVENT, PAST_EVENT]),
      updateEvent: () => of({ ...LIVE_EVENT, venue_name: 'New Venue' }),
    });
    const store = TestBed.inject(EventsStore);
    const refreshSpy = vi.spyOn(store, 'refresh');

    const fixture = TestBed.createComponent(EventsListComponent);
    (fixture.componentInstance as EventsListComponent).today = TODAY;
    fixture.detectChanges();

    const cmp = fixture.componentInstance as EventsListComponent;
    const row = cmp.rowsForActiveTab()[0];
    cmp.commitVenue(row, { value: 'New Venue', endValue: '' });

    expect(refreshSpy).toHaveBeenCalled();
  });
});

// Task 3: editable cells — commitVenue / commitDates / commitGenres call
// updateEvent() and the signal updates or reverts based on success/error.
describe('EventsListComponent — inline edits (Task 3)', () => {
  let fixture: ComponentFixture<EventsListComponent>;
  let component: EventsListComponent;
  let updateEventSpy: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    updateEventSpy = vi.fn().mockReturnValue(of(LIVE_EVENT));

    await TestBed.configureTestingModule({
      imports: [EventsListComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        {
          provide: ApiService,
          useValue: {
            getEvents: () => of([LIVE_EVENT, FUTURE_EVENT, PAST_EVENT]),
            updateEvent: updateEventSpy,
          },
        },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve() },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(EventsListComponent);
    component = fixture.componentInstance;
    component.today = TODAY;
    fixture.detectChanges();
  });

  it('renders events through the shared data-table', () => {
    const host: HTMLElement = fixture.nativeElement;
    expect(host.querySelector('app-data-table')).not.toBeNull();
    expect(host.textContent).toContain('Warehouse 9');
  });

  it('editing a venue cell PATCHes the event and updates the row', () => {
    const updatedEvent: Event = { ...LIVE_EVENT, venue_name: 'New Venue' };
    updateEventSpy.mockReturnValue(of(updatedEvent));
    // Stub out the store refresh so the synchronous mock getEvents() doesn't
    // reset the linkedSignal before we can assert the optimistic update.
    // (The refresh path is covered separately in the EL-091 cached-list suite.)
    vi.spyOn(TestBed.inject(EventsStore), 'refresh').mockImplementation(() => {});

    const cmp = fixture.componentInstance;
    const row = cmp.rowsForActiveTab()[0];
    cmp.commitVenue(row, { value: 'New Venue', endValue: '' });
    fixture.detectChanges();

    expect(updateEventSpy).toHaveBeenCalledWith(
      'evt-live',
      expect.objectContaining({ venue_name: 'New Venue' }),
    );
    expect(cmp.events().find((e) => e.id === 'evt-live')!.venue_name).toBe('New Venue');
  });

  it('reverts (leaves events unchanged) when the PATCH errors', () => {
    updateEventSpy.mockReturnValue(throwError(() => new Error('500')));

    const cmp = fixture.componentInstance;
    const row = cmp.rowsForActiveTab()[0];
    const before = cmp.events().find((e) => e.id === String(row['id']))!.venue_name;
    cmp.commitVenue(row, { value: 'Broken', endValue: '' });
    fixture.detectChanges();

    expect(cmp.events().find((e) => e.id === String(row['id']))!.venue_name).toBe(before);
  });

  it('commitDates sends start and end', () => {
    const updatedEvent: Event = { ...LIVE_EVENT, start_date: '2026-08-01', end_date: '2026-08-03' };
    updateEventSpy.mockReturnValue(of(updatedEvent));

    const cmp = fixture.componentInstance;
    const row = cmp.rowsForActiveTab()[0];
    cmp.commitDates(row, { value: '2026-08-01', endValue: '2026-08-03' });

    expect(updateEventSpy).toHaveBeenCalledWith(
      String(row['id']),
      expect.objectContaining({ start_date: '2026-08-01', end_date: '2026-08-03' }),
    );
  });

  it('commitGenres splits the comma string into an array', () => {
    const updatedEvent: Event = { ...LIVE_EVENT, genres: ['House', 'Techno'] };
    updateEventSpy.mockReturnValue(of(updatedEvent));

    const cmp = fixture.componentInstance;
    const row = cmp.rowsForActiveTab()[0];
    cmp.commitGenres(row, { value: 'House, Techno , ', endValue: '' });

    expect(updateEventSpy).toHaveBeenCalledWith(
      String(row['id']),
      expect.objectContaining({ genres: ['House', 'Techno'] }),
    );
  });
});
