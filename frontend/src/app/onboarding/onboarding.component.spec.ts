import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter, Router } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { signal } from '@angular/core';
import { OnboardingComponent } from './onboarding.component';
import { SessionStore } from '../services/session.store';
import { School } from '../services/api.service';

// ── helpers ──────────────────────────────────────────────────────────────────

const SCHOOL: School = { id: 's1', name: 'IMDJ School', role: 'owner' };
const MEMBER = { id: 'm1', email: 'dj@test.com', role: 'member', status: 'active' };

function makeStubSession(school: School | null = null) {
  return {
    school: signal<School | null>(school),
    user: signal(null),
    loading: signal(false),
    ensure: () => {},
    refresh: () => {},
  };
}

/** Cast nativeElement to HTMLElement and query. */
function qs<T extends Element>(el: unknown, sel: string): T | null {
  return (el as HTMLElement).querySelector<T>(sel);
}

// ── Shared describe — one configureTestingModule for the whole file ───────────

describe('OnboardingComponent', () => {
  let fixture: ComponentFixture<OnboardingComponent>;
  let httpMock: HttpTestingController;
  let router: Router;
  let stubSession: ReturnType<typeof makeStubSession>;

  // Configure once per test (Angular resets TestBed between tests).
  beforeEach(async () => {
    stubSession = makeStubSession(null);
    await TestBed.configureTestingModule({
      imports: [OnboardingComponent],
      providers: [
        { provide: SessionStore, useValue: stubSession },
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([
          { path: 'admin/events', children: [] },
          { path: 'onboarding', children: [] },
        ]),
        provideTranslateService(),
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(OnboardingComponent);
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  // ── Step 1 tests ───────────────────────────────────────────────────────────

  describe('step 1 — register school', () => {
    it('renders the school-name, school-city inputs and register button', () => {
      expect(qs(fixture.nativeElement, '[data-testid="school-name-input"]')).toBeTruthy();
      expect(qs(fixture.nativeElement, '[data-testid="school-city-input"]')).toBeTruthy();
      expect(qs(fixture.nativeElement, '[data-testid="register-school-btn"]')).toBeTruthy();
    });

    it('does not render step 2 inputs while on step 1', () => {
      expect(qs(fixture.nativeElement, '[data-testid="member-email-input"]')).toBeNull();
      expect(qs(fixture.nativeElement, '[data-testid="onboarding-done-btn"]')).toBeNull();
    });

    it('register-school-btn is disabled when school name is empty', () => {
      const btn = qs<HTMLButtonElement>(
        fixture.nativeElement,
        '[data-testid="register-school-btn"]',
      );
      expect(btn?.disabled).toBe(true);
    });

    it('register-school-btn is enabled after entering a school name', () => {
      const input = qs<HTMLInputElement>(
        fixture.nativeElement,
        '[data-testid="school-name-input"]',
      );
      input!.value = 'IMDJ School';
      input!.dispatchEvent(new Event('input'));
      fixture.detectChanges();
      const btn = qs<HTMLButtonElement>(
        fixture.nativeElement,
        '[data-testid="register-school-btn"]',
      );
      expect(btn?.disabled).toBe(false);
    });

    it('createSchool is called with name and city on submit', () => {
      const nameInput = qs<HTMLInputElement>(
        fixture.nativeElement,
        '[data-testid="school-name-input"]',
      );
      nameInput!.value = 'IMDJ School';
      nameInput!.dispatchEvent(new Event('input'));

      const cityInput = qs<HTMLInputElement>(
        fixture.nativeElement,
        '[data-testid="school-city-input"]',
      );
      cityInput!.value = 'Taipei';
      cityInput!.dispatchEvent(new Event('input'));
      fixture.detectChanges();

      qs<HTMLButtonElement>(fixture.nativeElement, '[data-testid="register-school-btn"]')!.click();
      fixture.detectChanges();

      const req = httpMock.expectOne((r) => r.url.endsWith('/api/schools'));
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ name: 'IMDJ School', city: 'Taipei' });
      req.flush(SCHOOL);
      // Flushing createSchool triggers _loadMembers() — drain that request too
      // so httpMock.verify() in afterEach doesn't throw.
      httpMock.expectOne((r) => r.url.endsWith('/api/schools/current/members')).flush([]);
    });

    it('calls session.refresh() and advances to step 2 after successful registration', () => {
      let refreshCalled = false;
      stubSession.refresh = () => {
        refreshCalled = true;
      };

      const nameInput = qs<HTMLInputElement>(
        fixture.nativeElement,
        '[data-testid="school-name-input"]',
      );
      nameInput!.value = 'IMDJ School';
      nameInput!.dispatchEvent(new Event('input'));
      fixture.detectChanges();

      qs<HTMLButtonElement>(fixture.nativeElement, '[data-testid="register-school-btn"]')!.click();
      fixture.detectChanges();

      httpMock.expectOne((r) => r.url.endsWith('/api/schools')).flush(SCHOOL);
      httpMock.expectOne((r) => r.url.endsWith('/api/schools/current/members')).flush([]);
      fixture.detectChanges();

      expect(refreshCalled).toBe(true);
      expect(qs(fixture.nativeElement, '[data-testid="onboarding-done-btn"]')).toBeTruthy();
    });
  });

  // ── Step 2 tests — advance via beforeEach (no second configureTestingModule) ─

  describe('step 2 — add team', () => {
    /** Advance from step 1 to step 2 within the same test setup. */
    function advanceToStep2() {
      const nameInput = qs<HTMLInputElement>(
        fixture.nativeElement,
        '[data-testid="school-name-input"]',
      );
      nameInput!.value = 'IMDJ School';
      nameInput!.dispatchEvent(new Event('input'));
      fixture.detectChanges();

      qs<HTMLButtonElement>(fixture.nativeElement, '[data-testid="register-school-btn"]')!.click();
      fixture.detectChanges();

      httpMock.expectOne((r) => r.url.endsWith('/api/schools')).flush(SCHOOL);
      httpMock.expectOne((r) => r.url.endsWith('/api/schools/current/members')).flush([MEMBER]);
      fixture.detectChanges();
    }

    it('renders step 2 inputs and done button', () => {
      advanceToStep2();
      expect(qs(fixture.nativeElement, '[data-testid="member-email-input"]')).toBeTruthy();
      expect(qs(fixture.nativeElement, '[data-testid="member-role-select"]')).toBeTruthy();
      expect(qs(fixture.nativeElement, '[data-testid="add-member-btn"]')).toBeTruthy();
      expect(qs(fixture.nativeElement, '[data-testid="onboarding-done-btn"]')).toBeTruthy();
    });

    it('renders the member list fetched from the backend', () => {
      advanceToStep2();
      const items = (fixture.nativeElement as HTMLElement).querySelectorAll('li');
      expect(items.length).toBeGreaterThanOrEqual(1);
      expect(items[0].textContent).toContain('dj@test.com');
    });

    it('inviteMember is called with email and role when add-member-btn is clicked', () => {
      advanceToStep2();

      // Set the signal directly (mirrors what ngModel would do on input change).
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (fixture.componentInstance as any).memberEmail.set('newdj@test.com');
      fixture.detectChanges();

      // The button should now be enabled and clicking it submits the form.
      const btn = qs<HTMLButtonElement>(fixture.nativeElement, '[data-testid="add-member-btn"]');
      expect(btn?.disabled).toBe(false);
      btn!.click();
      fixture.detectChanges();

      const req = httpMock.expectOne((r) => r.url.endsWith('/api/schools/current/members'));
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ email: 'newdj@test.com', role: 'member' });
      req.flush({ id: 'm2', email: 'newdj@test.com', role: 'member', status: 'pending' });
      fixture.detectChanges();

      // Email signal should reset to '' after success.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect((fixture.componentInstance as any).memberEmail()).toBe('');
    });

    it('done button navigates to /admin/events', () => {
      advanceToStep2();
      const navigateSpy = vi.spyOn(router, 'navigate');
      qs<HTMLButtonElement>(fixture.nativeElement, '[data-testid="onboarding-done-btn"]')!.click();
      expect(navigateSpy).toHaveBeenCalledWith(['/admin/events']);
    });
  });
});
