import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, Subject } from 'rxjs';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { DjsComponent } from './djs.component';
import { ApiService, DJ } from '../../services/api.service';
import { DjsStore } from '../../services/djs-store';
import { DialogService } from '../../shared/dialog.service';

// The DJs page must wrap itself in the console AdminShell so the left sidebar is
// always present and shows the DJs nav as active (it previously had no shell, so
// navigating to DJs dropped the sidebar entirely).
describe('DjsComponent — Console shell', () => {
  const apiMock = { getDJs: () => of([]) };

  function render() {
    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
      ],
    });
    const fixture = TestBed.createComponent(DjsComponent);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  it('wraps the page in the AdminShell with the DJs nav active', () => {
    const root = render();
    expect(root.querySelector('app-admin-shell')).toBeTruthy();
    expect(root.querySelector('.shell-sidebar')).toBeTruthy();
    expect(
      root.querySelector('[data-nav="djs"]')?.classList.contains('shell-nav-active'),
    ).toBe(true);
  });
});

// US-012 / EL-089: the organizer mints a DJ's portal link. Instead of a silent
// clipboard copy, it opens a dialog showing the link (so they can see what it is
// and copy it), carrying the URL and expiry for the regeneration warning.
describe('DjsComponent — portal link dialog (US-012 / EL-089)', () => {
  const shareLink = vi.fn();
  const apiMock = {
    getDJs: () => of([]),
    generateDJPortalToken: vi.fn(() =>
      of({
        portal_url: 'http://localhost:4200/dj/portal?token=tok',
        expires_at: '2026-07-17T00:00:00Z',
      })),
  };

  beforeEach(() => {
    apiMock.generateDJPortalToken.mockClear();
    shareLink.mockClear();

    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: DialogService, useValue: { shareLink } },
      ],
    });
  });

  it('mints a token and opens a share dialog carrying the portal URL', () => {
    const component = TestBed.createComponent(DjsComponent).componentInstance;

    component.copyPortalLink('dj-1');

    expect(apiMock.generateDJPortalToken).toHaveBeenCalledWith('dj-1');
    expect(shareLink).toHaveBeenCalledTimes(1);
    expect(shareLink.mock.calls[0][0]).toMatchObject({
      url: 'http://localhost:4200/dj/portal?token=tok',
    });
  });
});

// EL-020: the certifications edit panel.
describe('DjsComponent — certifications edit panel (EL-020)', () => {
  const DJ = { id: 'dj-1', name: 'Mia', genre_tags: ['House'], certifications: ['House'], is_student: true };
  const updateDJ = vi.fn(() => of({ ...DJ }));
  const apiMock = {
    getDJs: vi.fn(() => of([DJ])),
    updateDJ,
  };

  function makeComponent() {
    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
      ],
    });
    return TestBed.createComponent(DjsComponent).componentInstance;
  }

  beforeEach(() => {
    updateDJ.mockClear();
    apiMock.getDJs.mockClear();
  });

  it('openEdit loads the DJ into the form', () => {
    const c = makeComponent();
    c.openEdit({ id: 'dj-1' });
    expect(c.editing()?.id).toBe('dj-1');
    expect(c.editName()).toBe('Mia');
    expect(c.editIsStudent()).toBe(true);
    expect(c.editCerts()).toEqual(['House']);
  });

  it('toggleCert adds and removes a certification', () => {
    const c = makeComponent();
    c.openEdit({ id: 'dj-1' });
    c.toggleCert('Techno');
    expect(c.editCerts()).toContain('Techno');
    c.toggleCert('House');
    expect(c.editCerts()).not.toContain('House');
  });

  it('addCustomCert appends a free-text certification once', () => {
    const c = makeComponent();
    c.openEdit({ id: 'dj-1' });
    c.editCustom.set('Afrobeat');
    c.addCustomCert();
    expect(c.editCerts()).toContain('Afrobeat');
    expect(c.editCustom()).toBe('');
    // Adding the same value again is a no-op.
    c.editCustom.set('Afrobeat');
    c.addCustomCert();
    expect(c.editCerts().filter(x => x === 'Afrobeat')).toHaveLength(1);
  });

  it('saveEdit PATCHes the DJ with cert + student changes and closes the panel', () => {
    const c = makeComponent();
    c.openEdit({ id: 'dj-1' });
    c.editIsStudent.set(false);
    c.toggleCert('Hip Hop');
    c.saveEdit();

    expect(updateDJ).toHaveBeenCalledWith('dj-1', {
      name: 'Mia',
      genre_tags: ['House'],
      certifications: ['House', 'Hip Hop'],
      is_student: false,
    });
    expect(c.editing()).toBeNull();
  });
});

// EL-081: the records reskin renders certification status as icon + label status
// pills (Cleared / Pending) — never colour alone — so the roster passes a
// red/green colour-blindness check.
describe('DjsComponent — certification status pills (EL-081)', () => {
  function render(djs: unknown[]) {
    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: { getDJs: () => of(djs) } },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
      ],
    });
    const fixture = TestBed.createComponent(DjsComponent);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  it('renders a held certification as a "Cleared" pill carrying BOTH an icon glyph and the label', () => {
    const root = render([
      { id: 'dj-1', name: 'Mia', genre_tags: ['House'], certifications: ['House'], is_student: true },
    ]);
    const pill = root.querySelector('[data-testid="dj-cert-dj-1"]') as HTMLElement;
    expect(pill).toBeTruthy();
    expect(pill.classList).toContain('status-confirmed'); // Cleared semantics
    expect(pill.querySelector('svg.status-badge-glyph')).toBeTruthy(); // icon, not colour alone
    expect(pill.textContent).toContain('House'); // label
  });

  it('shows a "Pending" status pill (icon + label) for a student with no certifications', () => {
    const root = render([
      { id: 'dj-2', name: 'Ada', genre_tags: ['Techno'], certifications: [], is_student: true },
    ]);
    const pill = root.querySelector('[data-testid="dj-nocert-dj-2"]') as HTMLElement;
    expect(pill).toBeTruthy();
    expect(pill.classList).toContain('status-pending');
    expect(pill.querySelector('svg.status-badge-glyph')).toBeTruthy();
  });

  it('marks a graduate with an icon + label tier pill', () => {
    const root = render([
      { id: 'dj-3', name: 'Rex', genre_tags: [], certifications: [], is_student: false },
    ]);
    const pill = root.querySelector('.dj-grad-pill') as HTMLElement;
    expect(pill).toBeTruthy();
    expect(pill.querySelector('svg.dj-grad-glyph')).toBeTruthy();
  });
});

// EL-092: the roster is backed by the root DjsStore (stale-while-revalidate), so
// a re-visit paints from cache with no spinner and every roster mutation refreshes
// it — but generating a portal token (which doesn't change the roster) must not.
describe('DjsComponent — cached roster (EL-092)', () => {
  const ROSTER: DJ[] = [
    { id: 'd1', name: 'Mia', genre_tags: ['House'], certifications: ['House'], is_student: true },
  ];

  const configure = (api: Partial<ApiService>, dialog: Partial<DialogService> = {}) =>
    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: api },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve(), ...dialog },
        },
      ],
    });

  it('paints cached rows with NO spinner on a second visit (revalidates silently)', () => {
    const first = new Subject<DJ[]>();
    const second = new Subject<DJ[]>(); // stays pending — cache must cover the gap
    let calls = 0;
    configure({ getDJs: () => (++calls === 1 ? first : second).asObservable() });

    // First visit: nothing cached → spinner, no table.
    const f1 = TestBed.createComponent(DjsComponent);
    f1.detectChanges();
    const el1 = f1.nativeElement as HTMLElement;
    expect(el1.querySelector('app-loading-meter')).toBeTruthy();
    expect(el1.querySelector('app-data-table')).toBeFalsy();
    first.next(ROSTER);
    first.complete();
    f1.detectChanges();
    expect(el1.querySelector('app-data-table')).toBeTruthy();

    // Second visit shares the singleton store: cached rows show immediately even
    // though the revalidation request (`second`) is still in flight — no spinner.
    const f2 = TestBed.createComponent(DjsComponent);
    f2.detectChanges();
    const el2 = f2.nativeElement as HTMLElement;
    expect(el2.querySelector('app-loading-meter')).toBeFalsy();
    expect(el2.querySelector('app-data-table')).toBeTruthy();
  });

  it('refreshes the shared store after adding a DJ', async () => {
    configure({ getDJs: () => of(ROSTER), createDJ: () => of(ROSTER[0]) });
    const refreshSpy = vi.spyOn(TestBed.inject(DjsStore), 'refresh');

    const c = TestBed.createComponent(DjsComponent).componentInstance;
    c.newDJ.set({ name: 'Nova' });
    c.newDJGenres.set('Techno');
    await c.addDJ();

    expect(refreshSpy).toHaveBeenCalled();
  });

  it('refreshes the shared store after deleting a DJ', async () => {
    configure({ getDJs: () => of(ROSTER), deleteDJ: () => of({}) });
    const refreshSpy = vi.spyOn(TestBed.inject(DjsStore), 'refresh');

    const c = TestBed.createComponent(DjsComponent).componentInstance;
    await c.deleteDJ('d1');

    expect(refreshSpy).toHaveBeenCalled();
  });

  it('refreshes the shared store after saving cert edits', () => {
    configure({ getDJs: () => of(ROSTER), updateDJ: () => of(ROSTER[0]) });
    const refreshSpy = vi.spyOn(TestBed.inject(DjsStore), 'refresh');

    const c = TestBed.createComponent(DjsComponent).componentInstance;
    c.openEdit({ id: 'd1' });
    c.saveEdit();

    expect(refreshSpy).toHaveBeenCalled();
  });

  it('does NOT refresh the store when generating a portal link (roster unchanged)', () => {
    configure(
      { getDJs: () => of(ROSTER), generateDJPortalToken: () => of({ portal_url: 'https://x/p', expires_at: '' }) },
      { shareLink: vi.fn() },
    );
    const refreshSpy = vi.spyOn(TestBed.inject(DjsStore), 'refresh');

    const c = TestBed.createComponent(DjsComponent).componentInstance;
    c.copyPortalLink('d1');

    expect(refreshSpy).not.toHaveBeenCalled();
  });
});
