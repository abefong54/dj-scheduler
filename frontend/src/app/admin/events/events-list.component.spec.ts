import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of } from 'rxjs';
import { provideTranslateService } from '@ngx-translate/core';
import { vi } from 'vitest';

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

  it('renders the Actions column as icon-only buttons (edit, clone, delete)', () => {
    const liveRow = root().querySelector('[data-testid="event-card-evt-live"]')!;
    const editBtn = liveRow.querySelector('[data-testid="event-edit-evt-live"]');
    const cloneBtn = liveRow.querySelector('[data-testid="event-clone-evt-live"]');
    const deleteBtn = liveRow.querySelector('[data-testid="event-delete-evt-live"]');

    // All three actions present, each rendered as an icon (no visible text label).
    for (const btn of [editBtn, cloneBtn, deleteBtn]) {
      expect(btn).toBeTruthy();
      expect(btn!.querySelector('app-icon')).toBeTruthy();
      expect(btn!.querySelector('span')).toBeFalsy(); // iconOnly hides the label span
    }

    // Accessible name preserved on the icon-only buttons (title falls back to the label key).
    expect(editBtn!.getAttribute('title')).toBe('actions.edit');
    expect(cloneBtn!.getAttribute('title')).toBe('events.clone');
    expect(deleteBtn!.getAttribute('title')).toBe('actions.delete');
  });

  it('centers the Actions column header and cell', () => {
    expect(root().querySelector('.events-th-actions')).toBeTruthy();
    expect(root().querySelector('.events-td-actions')).toBeTruthy();
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
