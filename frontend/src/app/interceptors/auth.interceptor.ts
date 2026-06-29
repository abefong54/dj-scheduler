import { inject } from '@angular/core';
import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

/**
 * Attaches the organizer JWT as a Bearer header on outgoing requests, and on any
 * 401 response clears the token and redirects to /login (handles secret rotation
 * or expiry mid-session).
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  // The DJ portal is a separate, token-in-header auth context (EL-038): it carries
  // its own portal token in Authorization and isn't an organizer session. Leave
  // those requests untouched — don't overwrite the header, and don't treat a
  // portal 401 as an organizer-session expiry (which would bounce a DJ to /login).
  if (req.url.includes('/api/dj/portal')) {
    return next(req);
  }

  const token = auth.token();
  const authReq = token
    ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
    : req;

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
