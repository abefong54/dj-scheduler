import { Component } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { AdminShellComponent, ShellNav } from './admin-shell.component';

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
          { path: 'admin/schedule', children: [] },
        ]),
        provideTranslateService(),
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

    it('renders the four nav items (Events / DJs / Performance / Schedule)', () => {
      const items = root().querySelectorAll('.shell-nav-item');
      expect(items.length).toBe(4);
      const navIds = Array.from(items).map((el) => el.getAttribute('data-nav'));
      expect(navIds).toEqual(['events', 'djs', 'performance', 'schedule']);
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
      expect(navClasses('schedule')).not.toContain('shell-nav-active');
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
});
