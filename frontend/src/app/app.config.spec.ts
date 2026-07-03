// EL-096: motion & view-transition foundation. These guard that our router
// config stays valid after adding withViewTransitions(), and that the View
// Transitions API is actually invoked on navigation in this Angular version
// (the mechanism the "smooth between views" feel relies on). The real
// browser-level behaviour + token resolution is covered by the Playwright spec
// e2e/specs/view-transitions.spec.ts (jsdom has no startViewTransition).
import { Component } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router, RouterOutlet, provideRouter, withViewTransitions } from '@angular/router';

import { appConfig } from './app.config';

describe('appConfig (EL-096)', () => {
  it('still provides a working Router after adding view transitions', () => {
    TestBed.configureTestingModule({ providers: [...appConfig.providers] });
    expect(TestBed.inject(Router)).toBeTruthy();
  });
});

describe('router view transitions (EL-096)', () => {
  @Component({ standalone: true, template: 'a' })
  class PageA {}
  @Component({ standalone: true, template: 'b' })
  class PageB {}
  @Component({ standalone: true, imports: [RouterOutlet], template: '<router-outlet />' })
  class Host {}

  it('invokes the View Transitions API when the browser supports it', async () => {
    const startViewTransition = vi.fn((cb: () => void) => {
      cb();
      return {
        finished: Promise.resolve(),
        ready: Promise.resolve(),
        updateCallbackDone: Promise.resolve(),
        skipTransition: () => {},
      };
    });
    // jsdom has no startViewTransition; provide one so Angular's feature runs.
    (document as unknown as { startViewTransition: unknown }).startViewTransition = startViewTransition;

    TestBed.configureTestingModule({
      providers: [
        provideRouter(
          [
            { path: 'a', component: PageA },
            { path: 'b', component: PageB },
          ],
          withViewTransitions(),
        ),
      ],
    });
    const router = TestBed.inject(Router);
    const fixture = TestBed.createComponent(Host);
    fixture.detectChanges();

    await router.navigate(['/a']);
    fixture.detectChanges();
    await router.navigate(['/b']); // route→route change triggers a transition
    fixture.detectChanges();

    expect(startViewTransition).toHaveBeenCalled();
  });
});
