import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { Router, provideRouter } from '@angular/router';

import { authInterceptor } from './auth.interceptor';
import { AuthService } from '../services/auth.service';
import { StudentAuthService } from '../services/student-auth.service';

describe('authInterceptor — student vs organizer contexts', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let org: AuthService;
  let student: StudentAuthService;
  let router: Router;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
        provideRouter([]),
      ],
    });
    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
    org = TestBed.inject(AuthService);
    student = TestBed.inject(StudentAuthService);
    router = TestBed.inject(Router);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  it('attaches the STUDENT jwt (not the organizer jwt) on /api/student/* calls', () => {
    org.setToken('org-jwt');
    student.setToken('student-jwt');
    http.get('/api/student/challenges').subscribe();
    const req = httpMock.expectOne('/api/student/challenges');
    expect(req.request.headers.get('Authorization')).toBe('Bearer student-jwt');
    req.flush([]);
  });

  it('attaches the ORGANIZER jwt on teacher challenge routes (not /api/student/*)', () => {
    org.setToken('org-jwt');
    student.setToken('student-jwt');
    http.get('/api/challenges').subscribe();
    const req = httpMock.expectOne('/api/challenges');
    expect(req.request.headers.get('Authorization')).toBe('Bearer org-jwt');
    req.flush([]);
  });

  it('on a student 401, clears the student session and bounces to /student/login', () => {
    const nav = vi.spyOn(router, 'navigate').mockResolvedValue(true);
    student.setToken('student-jwt');
    http.get('/api/student/feed').subscribe({ next: () => {}, error: () => {} });
    httpMock
      .expectOne('/api/student/feed')
      .flush('nope', { status: 401, statusText: 'Unauthorized' });
    expect(student.token()).toBeNull();
    expect(nav).toHaveBeenCalledWith(['/student/login']);
  });
});
