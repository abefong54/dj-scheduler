import { Component, inject } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';
import { AuthService } from '../services/auth.service';

/**
 * EL-062 — Login restyled to the Console dark design language (design-system §7,
 * public/dark surface): a centered console card on a near-black violet field.
 *
 * Presentation only. The working sign-in path is still Google OAuth via
 * AuthService.signInWithGoogle(); the email / "magic link" control is rendered
 * visual-only and disabled because no magic-link endpoint exists in the backend.
 */
@Component({
  selector: 'app-login',
  standalone: true,
  imports: [TranslatePipe],
  template: `
    <section
      class="relative flex min-h-[calc(100dvh-3.5rem)] items-center justify-center overflow-hidden px-6 py-16"
      style="background:
        radial-gradient(55% 45% at 50% 40%, rgba(124, 58, 237, 0.20), transparent 70%),
        linear-gradient(180deg, var(--void) 0%, var(--void-2) 100%);">
      <div
        class="relative w-full max-w-sm rounded-[var(--radius-card)] border border-white/10 bg-white/[0.03]
               px-8 py-10 text-center shadow-[var(--shadow-lg)] backdrop-blur-sm">
        <h1 class="font-display text-3xl font-bold tracking-tight text-white">EventLineup</h1>

        <button
          type="button"
          (click)="login()"
          data-testid="google-signin"
          class="mt-8 flex w-full items-center justify-center gap-3 rounded-[var(--radius-input)]
                 bg-[var(--violet)] px-6 py-3 text-sm font-medium text-white transition-colors
                 hover:bg-[var(--violet-deep)] focus:outline-none focus-visible:ring-2
                 focus-visible:ring-[var(--cyan)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--void-2)]">
          <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
            <path fill="#fff" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62z"/>
            <path fill="#fff" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18z"/>
            <path fill="#fff" d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33z"/>
            <path fill="#fff" d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.58A9 9 0 0 0 .96 4.95L3.97 7.3C4.68 5.16 6.66 3.58 9 3.58z"/>
          </svg>
          {{ 'auth.continueWithGoogle' | translate }}
        </button>

        <div class="my-6 flex items-center gap-3 font-mono text-[10px] uppercase tracking-[0.18em] text-white/30">
          <span class="h-px flex-1 bg-white/10"></span>
          {{ 'auth.or' | translate }}
          <span class="h-px flex-1 bg-white/10"></span>
        </div>

        <!-- Visual-only: no magic-link endpoint exists. Disabled, never submits. -->
        <input
          type="email"
          disabled
          data-testid="magic-link-email"
          [attr.placeholder]="'auth.emailPlaceholder' | translate"
          class="w-full rounded-[var(--radius-input)] border border-[var(--violet)]/60 bg-[var(--violet)]/10
                 px-4 py-3 text-sm text-white shadow-[0_0_24px_-4px_rgba(124,58,237,0.6)]
                 placeholder:text-white/40 focus:outline-none disabled:cursor-not-allowed" />
        <button
          type="button"
          disabled
          data-testid="magic-link-submit"
          class="mt-3 w-full cursor-not-allowed rounded-[var(--radius-input)] bg-[var(--violet)]/40
                 px-6 py-3 text-sm font-medium text-white/60">
          {{ 'auth.sendMagicLink' | translate }}
        </button>
        <p class="mt-2 font-mono text-[10px] uppercase tracking-[0.16em] text-white/30">
          {{ 'auth.magicLinkComingSoon' | translate }}
        </p>

        <p class="mt-8 text-sm text-white/40">{{ 'auth.tagline' | translate }}</p>
      </div>
    </section>
  `,
})
export class LoginComponent {
  private auth = inject(AuthService);

  login(): void {
    this.auth.signInWithGoogle();
  }
}
