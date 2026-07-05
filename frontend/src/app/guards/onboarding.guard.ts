import { inject } from '@angular/core';
import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { toObservable } from '@angular/core/rxjs-interop';
import { Observable } from 'rxjs';
import { filter, map, take } from 'rxjs/operators';
import { SessionStore } from '../services/session.store';

/**
 * Guards admin routes (EL-106): ensures the session is loaded, then redirects
 * users who have no school yet to /onboarding so they can create one.
 * Users who already have a school pass through.
 */
export const onboardingGuard: CanActivateFn = (): boolean | UrlTree | Observable<boolean | UrlTree> => {
  const session = inject(SessionStore);
  const router = inject(Router);

  session.ensure();

  // Fast-path: session already resolved, no async needed.
  if (!session.loading()) {
    return session.school() === null
      ? router.createUrlTree(['/onboarding'])
      : true;
  }

  return toObservable(session.loading).pipe(
    filter((loading) => !loading),
    take(1),
    map(() =>
      session.school() === null
        ? router.createUrlTree(['/onboarding'])
        : true,
    ),
  );
};

/**
 * Guards the /onboarding route: users who already have a school are bounced
 * to /admin/events so they can't re-onboard.
 */
export const hasSchoolGuard: CanActivateFn = (): boolean | UrlTree | Observable<boolean | UrlTree> => {
  const session = inject(SessionStore);
  const router = inject(Router);

  session.ensure();

  if (!session.loading()) {
    return session.school() !== null
      ? router.createUrlTree(['/admin/events'])
      : true;
  }

  return toObservable(session.loading).pipe(
    filter((loading) => !loading),
    take(1),
    map(() =>
      session.school() !== null
        ? router.createUrlTree(['/admin/events'])
        : true,
    ),
  );
};
