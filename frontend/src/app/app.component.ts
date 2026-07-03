import { Component, inject } from '@angular/core';
import { NavigationEnd, Router, RouterLink, RouterOutlet } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { filter, map, startWith } from 'rxjs';
import { TranslatePipe } from '@ngx-translate/core';
import { LanguageService } from './services/language.service';
import { AuthService } from './services/auth.service';
import { ThemeService } from './services/theme.service';
import { DialogComponent } from './shared/dialog.component';
import { GlobalLoadingComponent } from './shared/global-loading.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, TranslatePipe, DialogComponent, GlobalLoadingComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  langService = inject(LanguageService);
  auth = inject(AuthService);
  // Instantiate the Soundcheck theme controller so the mode attribute stays in
  // sync with the app root (dark "booth" default; EL-078).
  private readonly theme = inject(ThemeService);

  private readonly router = inject(Router);
  readonly isLanding = toSignal(
    this.router.events.pipe(
      filter((e) => e instanceof NavigationEnd),
      map(() => this.onLandingRoute()),
      startWith(this.onLandingRoute()),
    ),
    { initialValue: this.onLandingRoute() },
  );
  private onLandingRoute(): boolean {
    return this.router.url.split(/[?#]/)[0] === '/';
  }
}
