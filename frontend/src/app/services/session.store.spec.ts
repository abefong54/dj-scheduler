import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { SessionStore } from './session.store';
import { MeResponse } from './api.service';

const ME_WITH_SCHOOL: MeResponse = {
  user: { id: 'u1', name: 'Abe', email: 'abe@test.com' },
  school: { id: 's1', name: 'IMDJ School', role: 'owner' },
};

const ME_NO_SCHOOL: MeResponse = {
  user: { id: 'u1', name: 'Abe', email: 'abe@test.com' },
  school: null,
};

describe('SessionStore', () => {
  let store: SessionStore;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [SessionStore, provideHttpClient(), provideHttpClientTesting()],
    });
    store = TestBed.inject(SessionStore);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('starts with null user, null school, and not loading', () => {
    expect(store.user()).toBeNull();
    expect(store.school()).toBeNull();
    expect(store.loading()).toBe(false);
  });

  it('ensure() fires /api/me and sets loading while in flight', () => {
    store.ensure();
    expect(store.loading()).toBe(true);
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_WITH_SCHOOL);
    expect(store.loading()).toBe(false);
    expect(store.user()).toEqual(ME_WITH_SCHOOL.user);
    expect(store.school()).toEqual(ME_WITH_SCHOOL.school);
  });

  it('ensure() is idempotent — a second call issues no additional request', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_WITH_SCHOOL);

    store.ensure(); // second call — should be a no-op
    httpMock.expectNone((r) => r.url.endsWith('/api/me'));
  });

  it('sets school to null when /api/me returns school: null', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_NO_SCHOOL);
    expect(store.school()).toBeNull();
    expect(store.user()).toEqual(ME_NO_SCHOOL.user);
  });

  it('refresh() always re-fetches regardless of prior ensure()', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_NO_SCHOOL);

    store.refresh();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_WITH_SCHOOL);
    expect(store.school()).toEqual(ME_WITH_SCHOOL.school);
  });

  it('refresh() can update the school after creation (null → school)', () => {
    store.ensure();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_NO_SCHOOL);
    expect(store.school()).toBeNull();

    store.refresh();
    httpMock.expectOne((r) => r.url.endsWith('/api/me')).flush(ME_WITH_SCHOOL);
    expect(store.school()).toEqual(ME_WITH_SCHOOL.school);
  });

  it('loading() returns false after a failed request', () => {
    store.ensure();
    httpMock
      .expectOne((r) => r.url.endsWith('/api/me'))
      .flush('server error', { status: 500, statusText: 'Server Error' });
    expect(store.loading()).toBe(false);
    // user/school stay null on error
    expect(store.user()).toBeNull();
    expect(store.school()).toBeNull();
  });
});
