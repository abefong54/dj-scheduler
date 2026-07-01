import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';

import { LandingComponent } from './landing.component';

describe('LandingComponent (EL-075 Soundcheck landing)', () => {
  async function build(): Promise<ComponentFixture<LandingComponent>> {
    await TestBed.configureTestingModule({
      imports: [LandingComponent],
      providers: [provideRouter([]), provideTranslateService()],
    }).compileComponents();
    const fixture = TestBed.createComponent(LandingComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('renders the text wordmark (no logo-SVG dependency)', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    const wordmarks = el.querySelectorAll('.lp-wordmark');
    // Quote block + footer both carry the wordmark.
    expect(wordmarks.length).toBeGreaterThanOrEqual(2);
    expect(wordmarks[0].textContent).toContain('soundcheck');
  });

  it('ports every major section from the comp', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    for (const sel of [
      '.lp-hero',
      '.lp-trust',
      '.lp-insight',
      '#how', // how-it-works
      '.lp-fgrid', // feature cards
      '.lp-statusband',
      '.lp-light .lp-records', // records + roster table
      '.lp-quote',
      '.lp-statsband',
      '.lp-cta-card',
      '.lp-footer',
    ]) {
      expect(el.querySelector(sel), `missing section: ${sel}`).toBeTruthy();
    }
  });

  it('renders all six feature cards and the three-state status system', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    expect(el.querySelectorAll('.lp-feat').length).toBe(6);
    expect(el.querySelectorAll('.lp-gcell').length).toBe(3);
  });

  it('renders the records roster table with every demo student', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    const rows = el.querySelectorAll('.lp-table tbody tr');
    expect(rows.length).toBe(4);
  });

  it('wires the primary CTAs to the sign-in / onboarding route', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    const primaries = el.querySelectorAll<HTMLAnchorElement>('a.lp-btn-amber');
    // Hero + CTA band both point at /login.
    expect(primaries.length).toBeGreaterThanOrEqual(2);
    primaries.forEach((a) => expect(a.getAttribute('href')).toBe('/login'));
  });

  it('uses translate keys (not raw copy) for section headings', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    // With no translations loaded, the pipe echoes the key — proves i18n wiring.
    expect(el.querySelector('.lp-h1')?.textContent).toContain('landing.hero.title');
  });
});
