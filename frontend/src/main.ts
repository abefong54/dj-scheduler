import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { AppComponent } from './app/app.component';
import { environment } from './environments/environment';

// EL-017 guard: fail loudly if a build shipped with apiUrl still set to the
// placeholder. netlify.toml rewrites `YOUR_RAILWAY_URL` to $RAILWAY_API_URL at
// build time; if that env var is unset the substitution is a no-op and every API
// call would silently hit a non-existent host. Throw before bootstrap so the
// misconfiguration is impossible to miss instead of surfacing as opaque failures.
if (environment.apiUrl.includes('YOUR_RAILWAY_URL')) {
  throw new Error(
    'environment.apiUrl still contains the YOUR_RAILWAY_URL placeholder — ' +
      'set the RAILWAY_API_URL build environment variable (see netlify.toml).',
  );
}

bootstrapApplication(AppComponent, appConfig)
  .catch((err) => console.error(err));
