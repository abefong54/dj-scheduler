import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';

import { LandingComponent } from './landing.component';

describe('LandingComponent (EL-079 Soundcheck reskin)', () => {
  async function build(): Promise<ComponentFixture<LandingComponent>> {
    await TestBed.configureTestingModule({
      imports: [LandingComponent],
      providers: [provideRouter([]), provideTranslateService()],
    }).compileComponents();
    const fixture = TestBed.createComponent(LandingComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('renders the page body on the dark-booth surface + ink tokens', async () => {
    const fixture = await build();
    const root = (fixture.nativeElement as HTMLElement).querySelector('div');
    expect(root?.getAttribute('style')).toContain('var(--surface)');
    expect(root?.getAttribute('style')).toContain('var(--ink)');
  });

  it('routes both CTAs to /login using the shared Cue-Amber .btn-primary', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const ctas = el.querySelectorAll<HTMLAnchorElement>('a[href="/login"]');
    expect(ctas.length).toBe(2);
    ctas.forEach((a) => {
      expect(a.className).toContain('btn');
      expect(a.className).toContain('btn-primary');
    });
  });

  it('drops the old light palette (no bg-white / slate wrappers remain)', async () => {
    const fixture = await build();
    const html = (fixture.nativeElement as HTMLElement).innerHTML;
    expect(html).not.toContain('bg-white');
    expect(html).not.toContain('text-slate-');
    expect(html).not.toContain('bg-slate-');
  });

  it('renders every feature card on the raised --card surface', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const today = el.querySelector('#today');
    const cards = today?.querySelectorAll('.grid > div');
    expect(cards?.length).toBe(6);
    cards?.forEach((c) => {
      expect(c.getAttribute('style')).toContain('var(--card)');
    });
  });
});
