import { ApplicationConfig } from '@angular/core';
import { provideRouter, withViewTransitions } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideTranslateService } from '@ngx-translate/core';
import { provideTranslateHttpLoader } from '@ngx-translate/http-loader';
import { routes } from './app.routes';
import { authInterceptor } from './interceptors/auth.interceptor';
import { loadingInterceptor } from './interceptors/loading.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    // EL-096: withViewTransitions() enables the browser View Transitions API so
    // route changes cross-fade instead of hard-cutting. Progressive enhancement —
    // unsupported browsers navigate as before. Timing/curve are set in styles.css
    // (::view-transition-*), honouring prefers-reduced-motion.
    provideRouter(routes, withViewTransitions()),
    provideHttpClient(withInterceptors([authInterceptor, loadingInterceptor])),
    // fallbackLang ensures that if the active language fails to load or is
    // missing a key, the UI degrades to English instead of showing raw keys.
    provideTranslateService({ lang: 'en', fallbackLang: 'en' }),
    ...provideTranslateHttpLoader({ prefix: '/assets/i18n/', suffix: '.json' }),
  ],
};
