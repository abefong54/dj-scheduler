import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Meta, Title } from '@angular/platform-browser';
import { TranslatePipe } from '@ngx-translate/core';

/**
 * Soundcheck marketing landing page (EL-075) — the branded front door prospects
 * see before signing in. Ported from the standalone "Backlit Precision" comp.
 *
 * Design notes:
 *  - All colour comes from the Soundcheck design tokens (styles.css `--sc-*` /
 *    semantic `--var()`), never the comp's hard-coded hexes. Layout + branded
 *    visuals live in landing.component.css (plain CSS, no @apply — stylelint).
 *  - Text wordmark only ("soundcheck" in the display face); it does NOT depend
 *    on the EL-071 logo SVGs.
 *  - Copy is wired through i18n (English). Traditional-Chinese product strings
 *    are DEFERRED to a native review (EL-070); missing zh-TW keys fall back to
 *    English via `fallbackLang`. Demo roster/name samples are literal sample
 *    data, not translated UI copy.
 */
@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterLink, TranslatePipe],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.css',
})
export class LandingComponent {
  /** Every primary CTA points at the sign-in / onboarding route. */
  readonly signInRoute = '/login';

  /** How-it-works steps — i18n under landing.how.<key>.{t,d}. */
  readonly steps = ['s1', 's2', 's3'] as const;

  /** Feature cards — i18n under landing.feat.<key>.{t,d}; accent tints the icon. */
  readonly features = [
    { key: 'credits', accent: 'amber' },
    { key: 'gates', accent: 'green' },
    { key: 'lineup', accent: 'cyan' },
    { key: 'cards', accent: 'amber' },
    { key: 'console', accent: 'cyan' },
    { key: 'taipei', accent: 'magenta' },
  ] as const;

  /** Status-system glyphs — i18n under landing.status.<key>.{n,s}. */
  readonly glyphs = ['cleared', 'pending', 'blocked'] as const;

  /** Records-table demo roster (sample data; names are proper nouns, not copy). */
  readonly roster = [
    { name: 'Yuki Kao', cjk: '高由紀', level: 'Level 3', credits: '18', certs: '4 / 5', tone: 'cleared' },
    { name: 'Marcus Lai', cjk: '', level: 'Level 3', credits: '22', certs: '5 / 5', tone: 'cleared' },
    { name: 'Sam Wu', cjk: '吳森', level: 'Level 2', credits: '9', certs: '3 / 4', tone: 'pending' },
    { name: 'Aria Chen', cjk: '陳雅', level: 'Level 1', credits: '3', certs: '1 / 3', tone: 'blocked' },
  ] as const;

  /** Stats band — value is sample data; caption is i18n under landing.stats.<key>. */
  readonly stats = [
    { key: 's1', value: '1,200+', accent: 'amber' },
    { key: 's2', value: '340', accent: 'cyan' },
    { key: 's3', value: '86%', accent: 'green' },
    { key: 's4', value: '3', accent: 'amber' },
  ] as const;

  /** Footer link columns — labels/links i18n under landing.footer.<col>.{title,<n>}. */
  readonly footerCols = [
    { key: 'product', links: [1, 2, 3, 4] },
    { key: 'for', links: [1, 2, 3, 4] },
    { key: 'company', links: [1, 2, 3] },
  ] as const;

  constructor() {
    const title = inject(Title);
    const meta = inject(Meta);

    // OG / social meta hooks. The share image asset lands in EL-077 — the path
    // here is the placeholder it will fill.
    const pageTitle = 'Soundcheck — performance platform for DJ schools';
    const description =
      'Soundcheck turns practice into stage time. Track every student’s reps, ' +
      'certify them stage-ready, and put cleared students on a real lineup.';
    const ogImage = '/assets/og/soundcheck-og.png';

    title.setTitle(pageTitle);
    meta.updateTag({ name: 'description', content: description });
    meta.updateTag({ property: 'og:type', content: 'website' });
    meta.updateTag({ property: 'og:title', content: pageTitle });
    meta.updateTag({ property: 'og:description', content: description });
    meta.updateTag({ property: 'og:image', content: ogImage });
    meta.updateTag({ name: 'twitter:card', content: 'summary_large_image' });
    meta.updateTag({ name: 'twitter:title', content: pageTitle });
    meta.updateTag({ name: 'twitter:description', content: description });
    meta.updateTag({ name: 'twitter:image', content: ogImage });
  }
}
