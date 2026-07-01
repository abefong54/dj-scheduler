import { Component, inject } from '@angular/core';
import { RouterOutlet, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { LanguageService } from './services/language.service';
import { AuthService } from './services/auth.service';
import { ThemeService } from './services/theme.service';
import { DialogComponent } from './shared/dialog.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, TranslatePipe, DialogComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  langService = inject(LanguageService);
  auth = inject(AuthService);
  // Instantiate the Soundcheck theme controller so the mode attribute stays in
  // sync with the app root (dark "booth" default; EL-078).
  private readonly theme = inject(ThemeService);
}
