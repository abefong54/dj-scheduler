import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ActivatedRoute, provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { of } from 'rxjs';

import { LandingComponent } from './landing.component';

describe('LandingComponent (EL-075 Soundcheck landing)', () => {
  async function build(): Promise<ComponentFixture<LandingComponent>> {
    await TestBed.configureTestingModule({
      imports: [LandingComponent],
      providers: [provideRouter([]), provideTranslateService(), provideHttpClient(), provideHttpClientTesting()],
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

  it('hero primary CTA opens the lead form (no /login link on the landing)', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('a[href="/login"]')).toBeNull();
    expect(el.querySelector('app-lead-form')).toBeNull();
    (el.querySelector('.lp-hero .lp-btn-amber') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(el.querySelector('app-lead-form')).not.toBeNull();
  });

  it('reveals the form when navigated with the #lead-form fragment', async () => {
    await TestBed.configureTestingModule({
      imports: [LandingComponent],
      providers: [
        provideRouter([]), provideTranslateService(), provideHttpClient(), provideHttpClientTesting(),
        { provide: ActivatedRoute, useValue: { fragment: of('lead-form') } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(LandingComponent);
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).querySelector('app-lead-form')).not.toBeNull();
  });

  it('uses translate keys (not raw copy) for section headings', async () => {
    const el = (await build()).nativeElement as HTMLElement;
    // With no translations loaded, the pipe echoes the key — proves i18n wiring.
    expect(el.querySelector('.lp-h1')?.textContent).toContain('landing.hero.title');
  });

  it('reveals the lead form when a demo CTA is clicked', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('app-lead-form')).toBeNull();
    const demoBtn = el.querySelector('.lp-cta-card .lp-btn-amber') as HTMLButtonElement;
    demoBtn.click();
    fixture.detectChanges();
    expect(el.querySelector('app-lead-form')).not.toBeNull();
  });
});
