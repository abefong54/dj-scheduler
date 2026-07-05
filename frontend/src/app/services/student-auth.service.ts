import { Injectable, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';
import { isExpired } from '../auth/jwt.util';

// Student session storage key. Deliberately DISTINCT from the organizer key
// ('jwt', see AuthService): a device may hold both an organizer and a student
// session, and the two auth contexts must never read or clobber each other.
const STUDENT_TOKEN_KEY = 'student_jwt';

/**
 * StudentAuthService is the student-scoped sibling of AuthService
 * (2026-07-05-student-google-login-design). It mirrors the organizer token model
 * exactly — the JWT lives in localStorage and is attached as a Bearer header by
 * the auth interceptor on `/api/student/*` calls — but keeps its own storage key
 * so student and organizer sessions stay isolated. The primary XSS mitigation is
 * the same host-layer CSP the organizer flow relies on (see AuthService for the
 * EL-040 storage rationale); migrating to an HttpOnly cookie is a shared follow-up.
 */
@Injectable({ providedIn: 'root' })
export class StudentAuthService {
  private router = inject(Router);

  private readonly _token = signal<string | null>(localStorage.getItem(STUDENT_TOKEN_KEY));

  /** The current student JWT, or null. */
  readonly token = this._token.asReadonly();

  /** True when a non-expired student token is present. */
  readonly isAuthenticated = computed(() => {
    const t = this._token();
    return t !== null && !isExpired(t);
  });

  setToken(token: string): void {
    localStorage.setItem(STUDENT_TOKEN_KEY, token);
    this._token.set(token);
  }

  clear(): void {
    localStorage.removeItem(STUDENT_TOKEN_KEY);
    this._token.set(null);
  }

  /**
   * Begin the student Google OAuth flow by navigating to the backend student
   * redirect route. Parallels AuthService.signInWithGoogle() but on the distinct
   * `/api/student/auth/google` route (the backend registers a second redirect URI
   * for it) so the callback issues a role=student JWT.
   */
  signInWithGoogle(): void {
    window.location.href = `${environment.apiUrl}/api/student/auth/google`;
  }

  signOut(): void {
    this.clear();
    this.router.navigate(['/student/login']);
  }
}
