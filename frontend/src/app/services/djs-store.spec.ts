import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { DjsStore } from './djs-store';
import { DJ } from './api.service';

const DJS: DJ[] = [
  { id: 'd1', name: 'Mia', genre_tags: ['House'], certifications: ['House'], is_student: true },
  { id: 'd2', name: 'Kenji', genre_tags: ['Techno'], certifications: [], is_student: false },
];

describe('DjsStore', () => {
  let store: DjsStore;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [DjsStore, provideHttpClient(), provideHttpClientTesting()],
    });
    store = TestBed.inject(DjsStore);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('exposes an empty roster and no spinner before the first load', () => {
    expect(store.djs()).toEqual([]);
    expect(store.loading()).toBe(false);
  });

  it('ensure() fetches /api/djs, showing the spinner until it resolves', () => {
    store.ensure();
    const req = httpMock.expectOne((r) => r.url.endsWith('/api/djs'));
    expect(store.loading()).toBe(true);
    req.flush(DJS);
    expect(store.loading()).toBe(false);
    expect(store.djs()).toEqual(DJS);
  });

  it('serves the cache with no spinner on a repeat ensure() (revalidates in background)', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/djs')).flush(DJS);

    store.ensure(); // second visit
    expect(store.loading()).toBe(false);
    expect(store.djs()).toEqual(DJS);
    httpMock.expectOne((r) => r.url.endsWith('/api/djs')).flush([DJS[0]]);
    expect(store.djs()).toEqual([DJS[0]]);
  });

  it('refresh() re-fetches the roster (invalidation after a mutation)', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/djs')).flush(DJS);

    store.refresh();
    httpMock.expectOne((r) => r.url.endsWith('/api/djs')).flush([DJS[1]]);
    expect(store.djs()).toEqual([DJS[1]]);
  });

  it('surfaces error() on a failed first load', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/djs'))
      .flush('nope', { status: 500, statusText: 'Server Error' });
    expect(store.error()).toBe(true);
    expect(store.loading()).toBe(false);
  });
});
