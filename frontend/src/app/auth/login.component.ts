import { Component, inject } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';
import { AuthService } from '../services/auth.service';

/**
 * EL-079 — Login reskinned onto the Soundcheck dark-"booth" system: a raised
 * Console Slate card (`--card`) on the Booth Black canvas (`--surface`), lit by a
 * faint Cue Amber glow. Supersedes the EL-062 Console (violet `--void`) surface;
 * all colour now flows from the app's semantic tokens (`--card`/`--ink`/`--muted`/
 * `--border`/`--violet`→Cue Amber/`--cyan`→Deck Cyan focus), which are re-pointed
 * to Soundcheck under `[data-brand="soundcheck"]` (active on <html>).
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
        radial-gradient(55% 45% at 50% 40%, rgba(255, 158, 44, 0.14), transparent 70%),
        var(--surface);">
      <div
        class="relative w-full max-w-sm rounded-[var(--radius-card)] border px-8 py-10 text-center"
        style="background: var(--card); border-color: var(--border); box-shadow: var(--shadow-lg);">
        <h1 class="font-display text-3xl font-bold tracking-tight" style="color: var(--ink);">EventLineup</h1>

        <!-- Primary action: reuse the shared Cue-Amber .btn-primary (dark ink +
             Deck-Cyan focus ring come from the token layer). -->
        <button
          type="button"
          (click)="login()"
          data-testid="google-signin"
          class="btn btn-primary mt-8 w-full gap-3 px-6 py-3 text-sm">
          <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true" fill="currentColor">
            <path d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62z"/>
            <path d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18z"/>
            <path d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33z"/>
            <path d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.58A9 9 0 0 0 .96 4.95L3.97 7.3C4.68 5.16 6.66 3.58 9 3.58z"/>
          </svg>
          {{ 'auth.continueWithGoogle' | translate }}
        </button>

        <div
          class="my-6 flex items-center gap-3 font-mono text-[10px] uppercase tracking-[0.18em]"
          style="color: var(--muted);">
          <span class="h-px flex-1" style="background: var(--border);"></span>
          {{ 'auth.or' | translate }}
          <span class="h-px flex-1" style="background: var(--border);"></span>
        </div>

        <!-- Visual-only: no magic-link endpoint exists. Disabled, never submits.
             Reuses the shared dark .form-input surface for the token look. -->
        <input
          type="email"
          disabled
          data-testid="magic-link-email"
          [attr.placeholder]="'auth.emailPlaceholder' | translate"
          class="form-input disabled:cursor-not-allowed disabled:opacity-60" />
        <button
          type="button"
          disabled
          data-testid="magic-link-submit"
          class="btn btn-secondary mt-3 w-full px-6 py-3 text-sm">
          {{ 'auth.sendMagicLink' | translate }}
        </button>
        <p class="mt-2 font-mono text-[10px] uppercase tracking-[0.16em]" style="color: var(--muted);">
          {{ 'auth.magicLinkComingSoon' | translate }}
        </p>

        <p class="mt-8 text-sm" style="color: var(--muted);">{{ 'auth.tagline' | translate }}</p>
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
