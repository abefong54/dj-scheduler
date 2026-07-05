import { Component } from '@angular/core';
import { environment } from '../../environments/environment';

/**
 * Student login (2026-07-05-student-google-login-design). SCAFFOLD shell: a
 * single "Sign in with Google" button that begins the student OAuth flow at the
 * backend student redirect route. Track E (login FE) fills the real presentation,
 * JWT storage on callback, and portal redirect.
 */
@Component({
  selector: 'app-student-login',
  standalone: true,
  template: `
    <section style="display:flex;min-height:60vh;align-items:center;justify-content:center;">
      <div style="text-align:center;">
        <h1>Student sign-in</h1>
        <p>SCAFFOLD — not implemented.</p>
        <button type="button" data-testid="student-google-signin" (click)="login()">
          Sign in with Google
        </button>
      </div>
    </section>
  `,
})
export class StudentLoginComponent {
  /** Begin the student Google OAuth flow at the backend redirect route. */
  login(): void {
    window.location.href = `${environment.apiUrl}/api/student/auth/google`;
  }
}
