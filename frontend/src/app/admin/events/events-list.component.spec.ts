import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { provideTranslateService } from '@ngx-translate/core';

import { EventsListComponent } from './events-list.component';
import { ApiService, Event } from '../../services/api.service';
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
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventsListComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
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
    const numbers = Array.from(root().querySelectorAll('.stat-tile-number')).map(
      el => el.textContent?.trim(),
    );
    // 2 upcoming (live + future), 1 past, 3 total.
    expect(numbers).toEqual(['2', '1', '3']);
  });

  it('renders a mono uppercase table header row', () => {
    const headers = root().querySelectorAll('.events-table .table-header');
    expect(headers.length).toBe(6);
  });

  it('renders one row per upcoming event with its data-testid', () => {
    expect(root().querySelector('[data-testid="event-card-evt-live"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-card-evt-future"]')).toBeTruthy();
    // Past event is on the Past tab, not the default Upcoming tab.
    expect(root().querySelector('[data-testid="event-card-evt-past"]')).toBeFalsy();
  });

  it('shows a LIVE status badge only for an in-progress event', () => {
    const liveRow = root().querySelector('[data-testid="event-card-evt-live"]');
    expect(liveRow?.querySelector('app-status-badge .status-live')).toBeTruthy();

    const futureRow = root().querySelector('[data-testid="event-card-evt-future"]');
    expect(futureRow?.querySelector('app-status-badge')).toBeFalsy();
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
    expect(root().querySelector('[data-testid="event-card-evt-past"]')).toBeTruthy();
    expect(root().querySelector('[data-testid="event-card-evt-live"]')).toBeFalsy();
  });

  it('derives lifecycle purely from dates', () => {
    expect(component.lifecycle(LIVE_EVENT)).toBe('live');
    expect(component.lifecycle(FUTURE_EVENT)).toBe('upcoming');
    expect(component.lifecycle(PAST_EVENT)).toBe('past');
    expect(component.isLive(LIVE_EVENT)).toBe(true);
    expect(component.isLive(FUTURE_EVENT)).toBe(false);
  });
});
