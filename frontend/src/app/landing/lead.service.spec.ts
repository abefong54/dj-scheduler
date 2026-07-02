import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { LeadService } from './lead.service';

describe('LeadService', () => {
  let svc: LeadService;
  let http: HttpTestingController;
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [LeadService, provideHttpClient(), provideHttpClientTesting()],
    });
    svc = TestBed.inject(LeadService);
    http = TestBed.inject(HttpTestingController);
  });
  afterEach(() => http.verify());

  it('POSTs the payload to /api/leads', () => {
    svc.submit({ name: 'A', organization: 'S', email: 'a@x.com', message: '' }).subscribe();
    const req = http.expectOne('/api/leads');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ name: 'A', organization: 'S', email: 'a@x.com', message: '' });
    req.flush({ id: '1' });
  });
});
