import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { DialogComponent } from './dialog.component';
import { DialogService } from './dialog.service';

// EL-089: the share dialog shows a copyable link. Clicking Copy writes the URL to
// the clipboard (the raw token only exists here — the DJs page no longer copies
// silently on its own).
describe('DialogComponent — share link (EL-089)', () => {
  let writeText: ReturnType<typeof vi.fn>;

  function render() {
    TestBed.configureTestingModule({
      imports: [DialogComponent],
      providers: [provideTranslateService(), DialogService],
    });
    const fixture = TestBed.createComponent(DialogComponent);
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => {
    writeText = vi.fn(() => Promise.resolve());
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
  });

  it('renders the URL and copies it to the clipboard on Copy', async () => {
    const fixture = render();
    const dialog = TestBed.inject(DialogService);

    dialog.shareLink({
      title: 'DJ schedule link',
      message: 'Send this to the DJ.',
      url: 'http://localhost:4200/dj/portal?token=tok',
      copyLabel: 'Copy link',
      copiedLabel: 'Copied!',
    });
    fixture.detectChanges();

    const input = fixture.nativeElement.querySelector(
      '[data-testid="dialog-link-input"]',
    ) as HTMLInputElement;
    expect(input.value).toBe('http://localhost:4200/dj/portal?token=tok');

    const copyBtn = fixture.nativeElement.querySelector(
      '[data-testid="dialog-link-copy"]',
    ) as HTMLButtonElement;
    copyBtn.click();
    await Promise.resolve();

    expect(writeText).toHaveBeenCalledWith('http://localhost:4200/dj/portal?token=tok');
  });
});
