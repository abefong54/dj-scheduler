import { Injectable, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';
import { isExpired } from '../auth/jwt.util';

const TOKEN_KEY = 'jwt';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private router = inject(Router);

  private readonly _token = signal<string | null>(localStorage.getItem(TOKEN_KEY));

  /** The current JWT, or null. */
  readonly token = this._token.asReadonly();

  /** True when a non-expired token is present. */
  readonly isAuthenticated = computed(() => {
    const t = this._token();
    return t !== null && !isExpired(t);
  });

  setToken(token: string): void {
    localStorage.setItem(TOKEN_KEY, token);
    this._token.set(token);
  }

  clear(): void {
    localStorage.removeItem(TOKEN_KEY);
    this._token.set(null);
  }

  /** Begin the Google OAuth flow by navigating to the backend redirect route. */
  signInWithGoogle(): void {
    window.location.href = `${environment.apiUrl}/auth/google`;
  }

  signOut(): void {
    this.clear();
    this.router.navigate(['/login']);
  }
}
