import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { signal } from '@angular/core';
import { Observable } from 'rxjs';
import { hasSchoolGuard, onboardingGuard } from './onboarding.guard';
import { SessionStore } from '../services/session.store';
import { School } from '../services/api.service';

// ── helpers ──────────────────────────────────────────────────────────────────

const SCHOOL: School = { id: 's1', name: 'IMDJ School', role: 'owner' };

/**
 * Stub SessionStore with loading=false so the guard takes its synchronous
 * fast-path — no async observable / fakeAsync needed.
 */
function makeStubSession(school: School | null) {
  return {
    school: signal<School | null>(school),
    user: signal(null),
    loading: signal(false), // already resolved — fast-path applies
    ensure: () => {},
    refresh: () => {},
  };
}

/**
 * Run a guard function inside an injection context and return the result.
 * The guard's fast-path (loading=false) returns synchronously, so no awaiting.
 */
function runGuard(guardFn: typeof onboardingGuard): boolean | UrlTree {
  return TestBed.runInInjectionContext(() =>
    guardFn({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot),
  ) as boolean | UrlTree;
}

// ── onboardingGuard ───────────────────────────────────────────────────────────

describe('onboardingGuard', () => {
  let router: Router;

  function setup(school: School | null) {
    TestBed.configureTestingModule({
      providers: [
        { provide: SessionStore, useValue: makeStubSession(school) },
        provideRouter([
          { path: 'onboarding', children: [] },
          { path: 'admin/events', children: [] },
        ]),
      ],
    });
    router = TestBed.inject(Router);
  }

  it('redirects to /onboarding when school is null', () => {
    setup(null);
    const result = runGuard(onboardingGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect(router.serializeUrl(result as UrlTree)).toBe('/onboarding');
  });

  it('returns true when the user has a school', () => {
    setup(SCHOOL);
    const result = runGuard(onboardingGuard);
    expect(result).toBe(true);
  });

  it('returns an Observable when loading is still in progress', () => {
    // Provide a session stub where loading is still true.
    TestBed.configureTestingModule({
      providers: [
        {
          provide: SessionStore,
          useValue: {
            school: signal(null),
            user: signal(null),
            loading: signal(true),
            ensure: () => {},
            refresh: () => {},
          },
        },
        provideRouter([{ path: 'onboarding', children: [] }]),
      ],
    });
    const result = TestBed.runInInjectionContext(() =>
      onboardingGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot),
    );
    // When loading=true the guard returns an Observable to wait for resolution.
    expect(result instanceof Observable).toBe(true);
  });
});

// ── hasSchoolGuard ────────────────────────────────────────────────────────────

describe('hasSchoolGuard', () => {
  let router: Router;

  function setup(school: School | null) {
    TestBed.configureTestingModule({
      providers: [
        { provide: SessionStore, useValue: makeStubSession(school) },
        provideRouter([
          { path: 'onboarding', children: [] },
          { path: 'admin/events', children: [] },
        ]),
      ],
    });
    router = TestBed.inject(Router);
  }

  it('redirects to /admin/events when the user already has a school', () => {
    setup(SCHOOL);
    const result = runGuard(hasSchoolGuard);
    expect(result).toBeInstanceOf(UrlTree);
    expect(router.serializeUrl(result as UrlTree)).toBe('/admin/events');
  });

  it('returns true (passes) when school is null', () => {
    setup(null);
    const result = runGuard(hasSchoolGuard);
    expect(result).toBe(true);
  });
});
