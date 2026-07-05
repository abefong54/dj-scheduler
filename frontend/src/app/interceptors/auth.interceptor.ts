import { inject } from '@angular/core';
import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { StudentAuthService } from '../services/student-auth.service';

/**
 * Attaches the organizer JWT as a Bearer header on outgoing requests, and on any
 * 401 response clears the token and redirects to /login (handles secret rotation
 * or expiry mid-session). Student-scoped `/api/student/*` calls use the separate
 * student JWT and bounce to /student/login instead — the two auth contexts are
 * isolated (a device may hold both).
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const studentAuth = inject(StudentAuthService);
  const router = inject(Router);

  // The DJ portal is a separate, token-in-header auth context (EL-038): it carries
  // its own portal token in Authorization and isn't an organizer session. Leave
  // those requests untouched — don't overwrite the header, and don't treat a
  // portal 401 as an organizer-session expiry (which would bounce a DJ to /login).
  if (req.url.includes('/api/dj/portal')) {
    return next(req);
  }

  // Student surfaces (2026-07-05 sprint): attach the student JWT and, on 401,
  // clear the student session and return to /student/login — never /login.
  if (req.url.includes('/api/student/')) {
    const studentToken = studentAuth.token();
    const studentReq = studentToken
      ? req.clone({ setHeaders: { Authorization: `Bearer ${studentToken}` } })
      : req;
    return next(studentReq).pipe(
      catchError((err: HttpErrorResponse) => {
        if (err.status === 401) {
          studentAuth.clear();
          router.navigate(['/student/login']);
        }
        return throwError(() => err);
      }),
    );
  }

  const token = auth.token();
  const authReq = token ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }) : req;

  return next(authReq).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401) {
        auth.clear();
        router.navigate(['/login']);
      }
      return throwError(() => err);
    }),
  );
};
