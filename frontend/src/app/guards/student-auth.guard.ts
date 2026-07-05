import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { StudentAuthService } from '../services/student-auth.service';

/**
 * Blocks student surfaces unless a valid, non-expired student JWT is present. An
 * expired or missing token is cleared and the student is sent to /student/login.
 * The student sibling of authGuard (2026-07-05-student-google-login-design).
 */
export const studentAuthGuard: CanActivateFn = () => {
  const auth = inject(StudentAuthService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }
  auth.clear();
  return router.createUrlTree(['/student/login']);
};
