import { TestBed } from '@angular/core/testing';
import {
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  UrlTree,
  provideRouter,
} from '@angular/router';

import { studentAuthGuard } from './student-auth.guard';
import { StudentAuthService } from '../services/student-auth.service';

function runGuard(): boolean | UrlTree {
  return TestBed.runInInjectionContext(() =>
    studentAuthGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot),
  ) as boolean | UrlTree;
}

describe('studentAuthGuard', () => {
  function setup(authenticated: boolean) {
    TestBed.configureTestingModule({
      providers: [
        {
          provide: StudentAuthService,
          useValue: { isAuthenticated: () => authenticated, clear: vi.fn() },
        },
        provideRouter([{ path: 'student/login', children: [] }]),
      ],
    });
  }

  it('allows navigation with a valid student session', () => {
    setup(true);
    expect(runGuard()).toBe(true);
  });

  it('redirects to /student/login without a session', () => {
    setup(false);
    const res = runGuard();
    expect(res).toBeInstanceOf(UrlTree);
    expect((res as UrlTree).toString()).toBe('/student/login');
  });
});
