import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { authInterceptor } from './auth.interceptor';
import { AuthService } from '../services/auth.service';

describe('authInterceptor', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let auth: AuthService;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
      ],
    });
    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
    auth = TestBed.inject(AuthService);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  it('attaches the organizer JWT to a normal API request', () => {
    auth.setToken('org-jwt');
    http.get('/api/events').subscribe();
    const req = httpMock.expectOne('/api/events');
    expect(req.request.headers.get('Authorization')).toBe('Bearer org-jwt');
    req.flush([]);
  });

  // EL-038: the DJ portal is a separate token-in-header auth context. Even if an
  // organizer happens to be logged in, the interceptor must not overwrite the
  // portal token with the organizer JWT.
  it('does not attach the organizer JWT to DJ portal requests', () => {
    auth.setToken('org-jwt');
    http.get('/api/dj/portal', { headers: { Authorization: 'Bearer portal-tok' } }).subscribe();
    const req = httpMock.expectOne('/api/dj/portal');
    expect(req.request.headers.get('Authorization')).toBe('Bearer portal-tok');
    req.flush({ dj: null, slots: [] });
  });
});
