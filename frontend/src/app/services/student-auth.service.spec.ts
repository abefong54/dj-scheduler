import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { StudentAuthService } from './student-auth.service';

// A syntactically valid JWT whose only meaningful claim is `exp`, base64url-encoded.
function tokenWithExp(secondsFromNow: number): string {
  const payload = { exp: Math.floor(Date.now() / 1000) + secondsFromNow };
  const b64 = btoa(JSON.stringify(payload))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
  return `header.${b64}.sig`;
}

describe('StudentAuthService', () => {
  let svc: StudentAuthService;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({ providers: [provideRouter([])] });
    svc = TestBed.inject(StudentAuthService);
  });

  afterEach(() => localStorage.clear());

  it('stores the token under a student-scoped key, isolated from the organizer key', () => {
    const t = tokenWithExp(3600);
    svc.setToken(t);
    expect(localStorage.getItem('student_jwt')).toBe(t);
    expect(localStorage.getItem('jwt')).toBeNull(); // never touches the organizer key
    expect(svc.token()).toBe(t);
    expect(svc.isAuthenticated()).toBe(true);
  });

  it('is not authenticated with an expired token', () => {
    svc.setToken(tokenWithExp(-10));
    expect(svc.isAuthenticated()).toBe(false);
  });

  it('clear removes the token', () => {
    svc.setToken(tokenWithExp(3600));
    svc.clear();
    expect(localStorage.getItem('student_jwt')).toBeNull();
    expect(svc.token()).toBeNull();
    expect(svc.isAuthenticated()).toBe(false);
  });
});
