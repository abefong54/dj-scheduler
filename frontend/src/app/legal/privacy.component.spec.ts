import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';

import { PrivacyComponent } from './privacy.component';

describe('PrivacyComponent (EL-112 privacy policy)', () => {
  async function build() {
    await TestBed.configureTestingModule({
      imports: [PrivacyComponent],
      providers: [provideRouter([]), provideTranslateService()],
    }).compileComponents();
    const fixture = TestBed.createComponent(PrivacyComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('renders the policy title and an email-usage section (what LINE reviews)', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    expect(el.querySelector('[data-testid="privacy-title"]')).toBeTruthy();
    // The email-usage section must be present — it's the crux of the LINE review.
    expect(el.querySelector('[data-testid="privacy-section-email"]')).toBeTruthy();
  });
});
