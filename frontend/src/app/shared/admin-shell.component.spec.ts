import { Component } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { signal } from '@angular/core';
import { AdminShellComponent, ShellNav } from './admin-shell.component';
import { SessionStore } from '../services/session.store';

/** Default stub: no school (covers the existing tests unchanged). */
function makeStubSession(schoolName: string | null = null) {
  return {
    school: signal(schoolName ? { id: 's1', name: schoolName, role: 'owner' } : null),
    user: signal(null),
    loading: signal(false),
    ensure: () => {},
    refresh: () => {},
  };
}

@Component({
  standalone: true,
  imports: [AdminShellComponent],
  template: `
    <app-admin-shell [activeNav]="active" [title]="title">
      <div class="projected-content">PROJECTED BODY</div>
    </app-admin-shell>
  `,
})
class HostComponent {
  active: ShellNav | null = 'events';
  title = 'My Page';
}

describe('AdminShellComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminShellComponent, HostComponent],
      providers: [
        // Stub routes for the nav links so clicking one (EL-093 close-on-nav)
        // resolves instead of throwing NG04002.
        provideRouter([
          { path: 'admin/events', children: [] },
          { path: 'admin/djs', children: [] },
          { path: 'admin/performance', children: [] },
        ]),
        provideTranslateService(),
        // EL-107: stub session so tests don't need HttpClient.
        { provide: SessionStore, useValue: makeStubSession(null) },
      ],
    }).compileComponents();
  });

  describe('chrome (via host with projected content)', () => {
    let fixture: ComponentFixture<HostComponent>;
    const root = () => fixture.nativeElement as HTMLElement;

    beforeEach(() => {
      fixture = TestBed.createComponent(HostComponent);
      fixture.detectChanges();
    });

    it('renders the dark sidebar opening at the WORKSPACE label (brand lives in the top bar, not the sidebar)', () => {
      expect(root().querySelector('.shell-sidebar')).toBeTruthy();
      // The wordmark was moved to the global top bar so the brand appears once.
      expect(root().querySelector('.shell-wordmark')).toBeNull();
      expect(
        root().querySelector('.shell-section-label')?.textContent,
      ).toContain('shell.workspace');
    });

    it('renders the white topbar with the page title from the title input', () => {
      expect(root().querySelector('.shell-topbar')).toBeTruthy();
      expect(root().querySelector('.shell-page-title')?.textContent).toContain(
        'My Page',
      );
    });

    it('renders the three nav items (Events / DJs / Performance)', () => {
      const items = root().querySelectorAll('.shell-nav-item');
      expect(items.length).toBe(3);
      const navIds = Array.from(items).map((el) => el.getAttribute('data-nav'));
      expect(navIds).toEqual(['events', 'djs', 'performance']);
    });

    it('does not render the removed Lineup/schedule nav item (dead /admin/schedule link)', () => {
      expect(root().querySelector('[data-nav="schedule"]')).toBeNull();
    });

    it('projects arbitrary page content into the light surface', () => {
      const content = root().querySelector('.shell-content .projected-content');
      expect(content?.textContent).toContain('PROJECTED BODY');
    });
  });

  describe('active-nav logic', () => {
    let fixture: ComponentFixture<AdminShellComponent>;
    const root = () => fixture.nativeElement as HTMLElement;
    const navClasses = (id: ShellNav) =>
      root().querySelector(`[data-nav="${id}"]`)!.classList;

    beforeEach(() => {
      fixture = TestBed.createComponent(AdminShellComponent);
    });

    it('marks only the active nav item with the active class', () => {
      fixture.componentRef.setInput('activeNav', 'events');
      fixture.detectChanges();
      expect(navClasses('events')).toContain('shell-nav-active');
      expect(navClasses('djs')).not.toContain('shell-nav-active');
      expect(navClasses('performance')).not.toContain('shell-nav-active');
    });

    it('moves the active class when activeNav changes', () => {
      fixture.componentRef.setInput('activeNav', 'events');
      fixture.detectChanges();
      fixture.componentRef.setInput('activeNav', 'djs');
      fixture.detectChanges();
      expect(navClasses('events')).not.toContain('shell-nav-active');
      expect(navClasses('djs')).toContain('shell-nav-active');
    });

    it('marks no item active when activeNav is null', () => {
      fixture.componentRef.setInput('activeNav', null);
      fixture.detectChanges();
      expect(root().querySelectorAll('.shell-nav-active').length).toBe(0);
    });
  });

  // EL-093: mobile drawer + overflow menu. Behaviour is asserted here; the
  // responsive layout itself (no overflow, reachability) is covered by the
  // Playwright spec e2e/specs/shell-mobile.spec.ts.
  describe('mobile drawer + overflow menu (EL-093)', () => {
    let fixture: ComponentFixture<AdminShellComponent>;
    const root = () => fixture.nativeElement as HTMLElement;
    const q = <T extends Element>(sel: string) => root().querySelector<T>(sel)!;

    beforeEach(() => {
      fixture = TestBed.createComponent(AdminShellComponent);
      fixture.detectChanges();
    });

    afterEach(() => {
      document.body.classList.remove('shell-drawer-lock');
    });

    it('drawer is closed by default', () => {
      expect(q('.shell-sidebar').classList).not.toContain('shell-sidebar-open');
      expect(q('.shell-hamburger').getAttribute('aria-expanded')).toBe('false');
    });

    it('the hamburger opens the drawer and locks body scroll', () => {
      q<HTMLButtonElement>('.shell-hamburger').click();
      fixture.detectChanges();
      expect(q('.shell-sidebar').classList).toContain('shell-sidebar-open');
      expect(q('.shell-hamburger').getAttribute('aria-expanded')).toBe('true');
      expect(document.body.classList).toContain('shell-drawer-lock');
    });

    it('clicking a nav item closes the drawer', () => {
      q<HTMLButtonElement>('.shell-hamburger').click();
      fixture.detectChanges();
      q<HTMLAnchorElement>('[data-nav="djs"]').click();
      fixture.detectChanges();
      expect(q('.shell-sidebar').classList).not.toContain('shell-sidebar-open');
      expect(document.body.classList).not.toContain('shell-drawer-lock');
    });

    it('clicking the backdrop closes the drawer', () => {
      q<HTMLButtonElement>('.shell-hamburger').click();
      fixture.detectChanges();
      q<HTMLElement>('.shell-backdrop').click();
      fixture.detectChanges();
      expect(q('.shell-sidebar').classList).not.toContain('shell-sidebar-open');
    });

    it('Escape closes the drawer', () => {
      q<HTMLButtonElement>('.shell-hamburger').click();
      fixture.detectChanges();
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
      fixture.detectChanges();
      expect(q('.shell-sidebar').classList).not.toContain('shell-sidebar-open');
    });

    it('the ⋯ toggle opens the actions group and stays open on its own click', () => {
      q<HTMLButtonElement>('.shell-actions-toggle').click();
      fixture.detectChanges();
      expect(q('.shell-actions-group').classList).toContain('open');
      expect(q('.shell-actions-toggle').getAttribute('aria-expanded')).toBe('true');
    });

    it('a click outside the actions group closes it', () => {
      q<HTMLButtonElement>('.shell-actions-toggle').click();
      fixture.detectChanges();
      document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }));
      fixture.detectChanges();
      expect(q('.shell-actions-group').classList).not.toContain('open');
    });

    it('Escape closes the actions group', () => {
      q<HTMLButtonElement>('.shell-actions-toggle').click();
      fixture.detectChanges();
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
      fixture.detectChanges();
      expect(q('.shell-actions-group').classList).not.toContain('open');
    });
  });

  // EL-100: the topbar search is opt-in — it only renders when a page actually
  // wires it (Event Detail). Everywhere else it was a dead input searching
  // nothing, so it must not render by default.
  describe('topbar search (opt-in)', () => {
    let fixture: ComponentFixture<AdminShellComponent>;
    const root = () => fixture.nativeElement as HTMLElement;

    beforeEach(() => {
      fixture = TestBed.createComponent(AdminShellComponent);
    });

    it('does not render the search input by default', () => {
      fixture.detectChanges();
      expect(root().querySelector('.shell-search')).toBeNull();
    });

    it('renders the search input when searchable is set', () => {
      fixture.componentRef.setInput('searchable', true);
      fixture.detectChanges();
      expect(root().querySelector('.shell-search')).toBeTruthy();
    });
  });

});

// ── EL-107: school name in sidebar ───────────────────────────────────────────
// Separate top-level describe so configureTestingModule doesn't run inside
// the existing describe's beforeEach and can control the SessionStore stub.

describe('AdminShellComponent — school name in sidebar (EL-107)', () => {
  it('shows the i18n fallback label when school is null', async () => {
    await TestBed.configureTestingModule({
      imports: [AdminShellComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: SessionStore, useValue: makeStubSession(null) },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(AdminShellComponent);
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('.shell-section-label')?.textContent?.trim(),
    ).toContain('shell.workspace');
  });

  it('shows the school name when session has a school', async () => {
    await TestBed.configureTestingModule({
      imports: [AdminShellComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: SessionStore, useValue: makeStubSession('IMDJ School') },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(AdminShellComponent);
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('.shell-section-label')?.textContent?.trim(),
    ).toContain('IMDJ School');
  });
});
