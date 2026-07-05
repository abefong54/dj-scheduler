import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { EventsStore } from './events-store';
import { Event } from './api.service';

const EVENTS: Event[] = [
  {
    id: 'e1',
    name: 'Summer Fest',
    venue_name: 'Hall A',
    start_date: '2026-07-01',
    end_date: '2026-07-02',
    genres: ['House'],
  },
  {
    id: 'e2',
    name: 'Winter Showcase',
    venue_name: 'Hall B',
    start_date: '2026-12-10',
    end_date: '2026-12-10',
    genres: ['Techno'],
  },
];

describe('EventsStore', () => {
  let store: EventsStore;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [EventsStore, provideHttpClient(), provideHttpClientTesting()],
    });
    store = TestBed.inject(EventsStore);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('exposes an empty list and no spinner before the first load', () => {
    expect(store.events()).toEqual([]);
    expect(store.loading()).toBe(false);
  });

  it('ensure() fetches /api/events, showing the spinner until it resolves', () => {
    store.ensure();
    const req = httpMock.expectOne((r) => r.url.endsWith('/api/events'));
    expect(store.loading()).toBe(true);
    req.flush(EVENTS);
    expect(store.loading()).toBe(false);
    expect(store.events()).toEqual(EVENTS);
  });

  it('serves the cache with no spinner on a repeat ensure() (revalidates in background)', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/events')).flush(EVENTS);

    store.ensure(); // second visit
    expect(store.loading()).toBe(false);
    expect(store.events()).toEqual(EVENTS);
    // a background revalidation request is still issued
    httpMock.expectOne((r) => r.url.endsWith('/api/events')).flush([EVENTS[0]]);
    expect(store.events()).toEqual([EVENTS[0]]);
  });

  it('refresh() re-fetches the list (invalidation after a mutation)', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/events')).flush(EVENTS);

    store.refresh();
    httpMock.expectOne((r) => r.url.endsWith('/api/events')).flush([EVENTS[1]]);
    expect(store.events()).toEqual([EVENTS[1]]);
  });

  it('surfaces error() on a failed first load', () => {
    store.ensure();
    httpMock
      .expectOne((r) => r.url.endsWith('/api/events'))
      .flush('nope', { status: 500, statusText: 'Server Error' });
    expect(store.error()).toBe(true);
    expect(store.loading()).toBe(false);
  });
});
