import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { DjsComponent } from './djs.component';
import { ApiService } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';

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
