import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { DjsComponent } from './djs.component';
import { ApiService } from '../../services/api.service';
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

// US-012: the organizer mints a DJ's portal link and copies it to the clipboard,
// with a brief "Copied!" flash.
describe('DjsComponent — copy portal link (US-012)', () => {
  let writeText: ReturnType<typeof vi.fn>;
  const apiMock = {
    getDJs: () => of([]),
    generateDJPortalToken: vi.fn(() =>
      of({ portal_url: 'http://localhost:4200/dj/portal?token=tok', expires_at: '' })),
  };

  beforeEach(() => {
    apiMock.generateDJPortalToken.mockClear();
    writeText = vi.fn(() => Promise.resolve());
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });

    TestBed.configureTestingModule({
      imports: [DjsComponent],
      providers: [
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
        { provide: DialogService, useValue: { confirm: () => Promise.resolve(true) } },
      ],
    });
  });

  it('mints a token, copies the URL to the clipboard, and flashes the DJ id', async () => {
    const component = TestBed.createComponent(DjsComponent).componentInstance;

    component.copyPortalLink('dj-1');
    // Flush the synchronous api emission and the async clipboard write callback.
    await Promise.resolve();
    await Promise.resolve();

    expect(apiMock.generateDJPortalToken).toHaveBeenCalledWith('dj-1');
    expect(writeText).toHaveBeenCalledWith('http://localhost:4200/dj/portal?token=tok');
    expect(component.copiedDJId()).toBe('dj-1');
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
