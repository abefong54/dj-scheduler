import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { StudentAuthService } from '../services/student-auth.service';

/**
 * Student login + Google-callback landing (2026-07-05-student-google-login-design,
 * Track E). The backend student OAuth callback redirects back here with either
 * `?token=<jwt>` on success or `?error=not_on_roster` when the Google account
 * isn't on any roster (enrollment is teacher-controlled). One component handles
 * both: store the token and forward to the portal, or show a friendly message.
 */
@Component({
  selector: 'app-student-login',
  standalone: true,
  template: `
    <section
      class="relative flex min-h-[calc(100dvh-3.5rem)] items-center justify-center overflow-hidden px-6 py-16"
      style="background:
        radial-gradient(55% 45% at 50% 40%, rgba(255, 158, 44, 0.14), transparent 70%),
        var(--surface);"
    >
      <div
        class="relative w-full max-w-sm rounded-[var(--radius-card)] border px-8 py-10 text-center"
        style="background: var(--card); border-color: var(--border); box-shadow: var(--shadow-lg);"
      >
        <h1 class="font-display text-3xl font-bold tracking-tight" style="color: var(--ink);">
          Student sign-in
        </h1>
        <p class="mt-3 text-sm" style="color: var(--muted);">
          Use the Google account your teacher has on file.
        </p>

        @if (error()) {
          <p
            data-testid="student-login-error"
            class="mt-6 rounded-[var(--radius-card)] border px-4 py-3 text-sm"
            style="color: var(--ink); border-color: var(--border); background: rgba(220, 38, 38, 0.08);"
          >
            {{ error() }}
          </p>
        }

        <button
          type="button"
          (click)="signIn()"
          data-testid="student-google-signin"
          class="btn btn-primary mt-8 w-full gap-3 px-6 py-3 text-sm"
        >
          <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true" fill="currentColor">
            <path
              d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62z"
            />
            <path
              d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18z"
            />
            <path d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33z" />
            <path
              d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.58A9 9 0 0 0 .96 4.95L3.97 7.3C4.68 5.16 6.66 3.58 9 3.58z"
            />
          </svg>
          Sign in with Google
        </button>
      </div>
    </section>
  `,
})
export class StudentLoginComponent {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(StudentAuthService);

  readonly error = signal<string | null>(null);

  constructor() {
    const qp = this.route.snapshot.queryParamMap;

    const token = qp.get('token');
    if (token) {
      this.auth.setToken(token);
      // Strip the JWT from the URL so it never lingers in history or a Referer.
      history.replaceState(history.state, '', window.location.pathname);
      this.router.navigate(['/student/challenges']);
      return;
    }

    if (qp.get('error') === 'not_on_roster') {
      this.error.set(
        "You're not on a roster yet — ask your teacher to add your email, then try signing in again.",
      );
      history.replaceState(history.state, '', window.location.pathname);
      return;
    }

    // Already signed in? Skip straight to the portal.
    if (this.auth.isAuthenticated()) {
      this.router.navigate(['/student/challenges']);
    }
  }

  signIn(): void {
    this.auth.signInWithGoogle();
  }
}
