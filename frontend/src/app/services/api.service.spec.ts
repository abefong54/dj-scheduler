import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ApiService, Slot, PublicSlot } from './api.service';

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
  dj_confirmation: null,
};

describe('ApiService', () => {
  let api: ApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ApiService, provideHttpClient(), provideHttpClientTesting()],
    });
    api = TestBed.inject(ApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  // EL-038: the DJ portal token must travel in the Authorization header, never
  // a URL query param, so it can't leak via history/Referer/logs.
  describe('DJ portal token delivery', () => {
    it('getDJPortal sends the token as a Bearer header, not a query param', () => {
      api.getDJPortal('portal-tok-123').subscribe();

      const req = httpMock.expectOne(r => r.url === '/api/dj/portal');
      expect(req.request.headers.get('Authorization')).toBe('Bearer portal-tok-123');
      expect(req.request.params.get('token')).toBeNull();
      expect(req.request.urlWithParams).not.toContain('token=');
      req.flush({ dj: null, slots: [] });
    });

    it('confirmSlot sends the token as a Bearer header, not a query param', () => {
      api.confirmSlot('slot-1', 'confirmed', 'portal-tok-456').subscribe();

      const req = httpMock.expectOne(r => r.url === '/api/dj/portal/slots/slot-1');
      expect(req.request.method).toBe('PATCH');
      expect(req.request.headers.get('Authorization')).toBe('Bearer portal-tok-456');
      expect(req.request.params.get('token')).toBeNull();
      expect(req.request.urlWithParams).not.toContain('token=');
      req.flush(SLOT);
    });
  });

  describe('getPublicSlot (EL-049)', () => {
    // Verify the method+path the backend actually registers (GET
    // /api/slots/:id/public) — the EL-035 contract guard for the share card.
    it('GETs the public single-slot endpoint and returns {slot, event}', () => {
      const payload: PublicSlot = { slot: SLOT, event: { id: 'evt-1', name: 'Spring Showcase', venue_name: 'Revolver', start_date: '2026-07-01', end_date: '2026-07-01', genres: [] } };

      let result: PublicSlot | undefined;
      api.getPublicSlot('slot-1').subscribe(r => (result = r));

      const req = httpMock.expectOne('/api/slots/slot-1/public');
      expect(req.request.method).toBe('GET');
      // Public endpoint — must not carry an Authorization header.
      expect(req.request.headers.get('Authorization')).toBeNull();

      req.flush(payload);
      expect(result).toEqual(payload);
    });
  });

  // US-012: organizer mints a DJ's portal link to hand out.
  describe('generateDJPortalToken', () => {
    it('POSTs to the DJ token route and returns the portal URL', () => {
      let result: { portal_url: string; expires_at: string } | undefined;
      api.generateDJPortalToken('dj-9').subscribe(r => (result = r));

      const req = httpMock.expectOne('/api/djs/dj-9/token');
      expect(req.request.method).toBe('POST');

      const body = { portal_url: 'http://localhost:4200/dj/portal?token=abc', expires_at: '2026-10-01T00:00:00Z' };
      req.flush(body);
      expect(result).toEqual(body);
    });
  });

  describe('updateSlot', () => {
    // Regression for EL-035: the backend registers PATCH (not PUT) for the
    // single-slot route, so a PUT here 404s and slot edits silently fail.
    it('issues a PATCH (not PUT) to the single-slot route with the slot body', () => {
      const body = {
        stage_id: SLOT.stage_id,
        dj_id: SLOT.dj_id,
        genre: SLOT.genre,
        slot_date: SLOT.slot_date,
        start_time: SLOT.start_time,
        end_time: SLOT.end_time,
        notes: SLOT.notes,
      };

      let result: Slot | undefined;
      api.updateSlot('evt-1', 'slot-1', body).subscribe(s => (result = s));

      const req = httpMock.expectOne('/api/events/evt-1/slots/slot-1');
      expect(req.request.method).toBe('PATCH');
      expect(req.request.body).toEqual(body);

      req.flush(SLOT);
      expect(result).toEqual(SLOT);
    });
  });

  // EL-043/044: verify the exact method+path the backend registers for the three
  // performance endpoints (the EL-035 contract guard).
  describe('performance endpoints (EL-043/044)', () => {
    it('getPerformance GETs /api/performance', () => {
      api.getPerformance().subscribe();
      const req = httpMock.expectOne('/api/performance');
      expect(req.request.method).toBe('GET');
      req.flush([]);
    });

    it('getPerformance forwards event_id/from/to as query params', () => {
      api.getPerformance({ eventId: 'e1', from: '2026-07-01', to: '2026-07-31' }).subscribe();
      const req = httpMock.expectOne(r => r.url === '/api/performance');
      expect(req.request.params.get('event_id')).toBe('e1');
      expect(req.request.params.get('from')).toBe('2026-07-01');
      expect(req.request.params.get('to')).toBe('2026-07-31');
      req.flush([]);
    });

    it('getUnderserved GETs /api/performance/underserved and passes threshold', () => {
      api.getUnderserved(3).subscribe();
      const req = httpMock.expectOne(r => r.url === '/api/performance/underserved');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('threshold')).toBe('3');
      req.flush([]);
    });

    it('getUnderserved omits threshold when not provided', () => {
      api.getUnderserved().subscribe();
      const req = httpMock.expectOne('/api/performance/underserved');
      expect(req.request.params.get('threshold')).toBeNull();
      req.flush([]);
    });

    it('getDJPerformance GETs /api/djs/:id/performance', () => {
      api.getDJPerformance('dj-1').subscribe();
      const req = httpMock.expectOne('/api/djs/dj-1/performance');
      expect(req.request.method).toBe('GET');
      req.flush({ dj_id: 'dj-1', dj_name: 'X', reps: 0, total_minutes: 0, last_played: '', by_genre: [] });
    });
  });
});
