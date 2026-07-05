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
    expect(root.querySelector('[data-nav="djs"]')?.classList.contains('shell-nav-active')).toBe(
      true,
    );
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
      }),
    ),
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
  const DJ = {
    id: 'dj-1',
    name: 'Mia',
    genre_tags: ['House'],
    certifications: ['House'],
    is_student: true,
  };
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
    expect(c.editCerts().filter((x) => x === 'Afrobeat')).toHaveLength(1);
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
      email: '',
    });
    expect(c.editing()).toBeNull();
  });

  it('saveEdit sends the roster email so the student can sign in', () => {
    const c = makeComponent();
    c.openEdit({ id: 'dj-1' });
    c.editEmail.set('mia@example.com');
    c.saveEdit();
    expect(updateDJ).toHaveBeenCalledWith(
      'dj-1',
      expect.objectContaining({ email: 'mia@example.com' }),
    );
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
      {
        id: 'dj-1',
        name: 'Mia',
        genre_tags: ['House'],
        certifications: ['House'],
        is_student: true,
      },
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
          useValue: {
            confirm: () => Promise.resolve(true),
            alert: () => Promise.resolve(),
            ...dialog,
          },
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
      {
        getDJs: () => of(ROSTER),
        generateDJPortalToken: () => of({ portal_url: 'https://x/p', expires_at: '' }),
      },
      { shareLink: vi.fn() },
    );
    const refreshSpy = vi.spyOn(TestBed.inject(DjsStore), 'refresh');

    const c = TestBed.createComponent(DjsComponent).componentInstance;
    c.copyPortalLink('d1');

    expect(refreshSpy).not.toHaveBeenCalled();
  });
});

// EL-099: the "Add a DJ" action moves into the topbar (next to search, like the
// Events "+ New") and opens the Name+Genres form in a modal — replacing the
// always-visible inline add card. The modal reuses .modal-overlay/.modal-box-sm
// so it inherits the shared present-animation (EL-096 motion foundation).
describe('DjsComponent — add via topbar modal (EL-099)', () => {
  const ROSTER: DJ[] = [
    { id: 'd1', name: 'Mia', genre_tags: ['House'], certifications: ['House'], is_student: true },
  ];
  const createDJ = vi.fn(() => of(ROSTER[0]));

  const setup = (api: Partial<ApiService> = {}) => {
    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: ApiService, useValue: { getDJs: () => of(ROSTER), createDJ, ...api } },
        {
          provide: DialogService,
          useValue: { confirm: () => Promise.resolve(true), alert: () => Promise.resolve() },
        },
      ],
    });
    const fixture = TestBed.createComponent(DjsComponent);
    fixture.detectChanges();
    return fixture;
  };

  beforeEach(() => createDJ.mockClear());

  it('shows no inline add form on load — only a topbar "+ Add" button', () => {
    const el = setup().nativeElement as HTMLElement;
    expect(el.querySelector('[data-testid="dj-name-input"]')).toBeNull(); // inline card gone
    expect(el.querySelector('.modal-overlay')).toBeNull(); // modal closed
    expect(el.querySelector('.shell-new-btn')).toBeTruthy(); // topbar add button
  });

  it('opens the add modal (reusing the shared modal shell) when the topbar button is clicked', () => {
    const fixture = setup();
    const el = fixture.nativeElement as HTMLElement;
    (el.querySelector('.shell-new-btn') as HTMLButtonElement).click();
    fixture.detectChanges();
    const modal = el.querySelector('.modal-overlay');
    expect(modal).toBeTruthy();
    // uses the shared animated box so it inherits the EL-096 present-animation
    expect(modal!.querySelector('.modal-box-sm')).toBeTruthy();
    expect(modal!.querySelector('[data-testid="dj-name-input"]')).toBeTruthy();
    expect(modal!.querySelector('[data-testid="dj-genres-input"]')).toBeTruthy();
  });

  it('adds a DJ from the modal and closes it', () => {
    const fixture = setup();
    const el = fixture.nativeElement as HTMLElement;
    const c = fixture.componentInstance;
    c.openAddModal();
    c.newDJ.set({ name: 'Nova' });
    c.newDJGenres.set('Techno');
    fixture.detectChanges();
    (el.querySelector('[data-testid="dj-add-btn"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(createDJ).toHaveBeenCalledWith({ name: 'Nova', genre_tags: ['Techno'] });
    expect(el.querySelector('.modal-overlay')).toBeNull(); // closed on success
  });

  it('cancel closes the modal without creating a DJ', () => {
    const fixture = setup();
    const el = fixture.nativeElement as HTMLElement;
    fixture.componentInstance.openAddModal();
    fixture.detectChanges();
    (el.querySelector('[data-testid="dj-add-cancel"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(el.querySelector('.modal-overlay')).toBeNull();
    expect(createDJ).not.toHaveBeenCalled();
  });

  // The add/cancel buttons must use the shared app-button style like every other
  // modal (edit panel, confirm dialog) — i.e. carry the `.btn` base class + a
  // variant. Raw `class="btn-primary"` buttons missed the base class, so they
  // rendered unstyled and didn't match the rest of the app.
  it('renders the add/cancel buttons with the shared app-button style (btn base + variant)', () => {
    const fixture = setup();
    const el = fixture.nativeElement as HTMLElement;
    fixture.componentInstance.openAddModal();
    fixture.detectChanges();

    const add = el.querySelector('[data-testid="dj-add-btn"]') as HTMLButtonElement;
    const cancel = el.querySelector('[data-testid="dj-add-cancel"]') as HTMLButtonElement;

    expect(add.classList.contains('btn')).toBe(true);
    expect(add.classList.contains('btn-primary')).toBe(true);
    expect(cancel.classList.contains('btn')).toBe(true);
    expect(cancel.classList.contains('btn-secondary')).toBe(true);
  });
});
