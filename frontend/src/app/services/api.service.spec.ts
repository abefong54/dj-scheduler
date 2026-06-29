import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ApiService, Slot } from './api.service';

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
});
