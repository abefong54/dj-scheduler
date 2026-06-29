import { Component, inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

/**
 * Lands here after Google redirects back via the backend with ?token=<jwt>.
 * Stores the token and forwards to the admin area, or to /login on failure.
 */
@Component({
  selector: 'app-auth-callback',
  standalone: true,
  template: `<p class="p-8 text-center text-gray-500">Signing you in…</p>`,
})
export class AuthCallbackComponent {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);

  constructor() {
    const token = this.route.snapshot.queryParamMap.get('token');
    if (token) {
      this.auth.setToken(token);
      // EL-038: strip the JWT from the URL before navigating, so it never lingers
      // in the address bar, browser history, or a Referer header.
      history.replaceState(history.state, '', window.location.pathname);
      this.router.navigate(['/admin/events']);
    } else {
      this.router.navigate(['/login']);
    }
  }
}
