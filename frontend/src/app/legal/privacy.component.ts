import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

/**
 * Public Privacy Policy page (EL-112 support). Linked from the landing footer and
 * the login consent line; its screenshot backs the LINE Login email-permission
 * review. Content is i18n (en + zh-TW) and driven by the `sections` list.
 */
@Component({
  selector: 'app-privacy',
  standalone: true,
  imports: [TranslatePipe, RouterLink],
  template: `
    <section class="min-h-[100dvh] px-6 py-16" style="background: var(--surface); color: var(--ink);">
      <main class="mx-auto max-w-2xl">
        <a
          routerLink="/"
          class="font-mono text-[11px] uppercase tracking-[0.16em]"
          style="color: var(--muted);"
          >← {{ 'privacy.back' | translate }}</a
        >
        <h1 class="font-display mt-6 text-3xl font-bold tracking-tight" data-testid="privacy-title">
          {{ 'privacy.title' | translate }}
        </h1>
        <p class="mt-2 font-mono text-[11px] uppercase tracking-[0.14em]" style="color: var(--muted);">
          {{ 'privacy.updated' | translate }}
        </p>
        <p class="mt-6 leading-relaxed" style="color: var(--muted);">{{ 'privacy.intro' | translate }}</p>

        @for (s of sections; track s) {
          <section class="mt-8" [attr.data-testid]="'privacy-section-' + s">
            <h2 class="text-lg font-semibold">{{ 'privacy.' + s + '.title' | translate }}</h2>
            <p class="mt-2 leading-relaxed" style="color: var(--muted);">
              {{ 'privacy.' + s + '.body' | translate }}
            </p>
          </section>
        }
      </main>
    </section>
  `,
})
export class PrivacyComponent {
  /** Policy sections; each maps to privacy.<key>.{title,body} i18n keys. */
  readonly sections = ['collect', 'email', 'share', 'retention', 'contact'] as const;
}
