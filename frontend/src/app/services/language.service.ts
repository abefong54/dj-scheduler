import { Injectable, signal, effect } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

/** Languages that ship a translation file under assets/i18n. */
export const SUPPORTED_LANGS = ['en', 'zh-TW'] as const;
export type SupportedLang = (typeof SUPPORTED_LANGS)[number];
const DEFAULT_LANG: SupportedLang = 'en';

@Injectable({ providedIn: 'root' })
export class LanguageService {
  currentLang = signal<string>(DEFAULT_LANG);

  constructor(private translate: TranslateService) {
    this.initLanguage();
    this.setupLanguageEffect();
  }

  private initLanguage() {
    // Coerce the persisted value to a supported language. A stale/legacy code
    // (e.g. "zh" from an earlier build) would 404 its translation file and,
    // with no fallback, leave the entire UI rendering raw keys. The effect
    // below rewrites localStorage with the normalized value, healing it.
    this.currentLang.set(normalizeLang(localStorage.getItem('lang')));
  }

  private setupLanguageEffect() {
    effect(() => {
      const lang = this.currentLang();
      this.translate.use(lang);
      localStorage.setItem('lang', lang);
    });
  }

  setLanguage(lang: string) {
    this.currentLang.set(normalizeLang(lang));
  }
}

function normalizeLang(lang: string | null): SupportedLang {
  return (SUPPORTED_LANGS as readonly string[]).includes(lang ?? '')
    ? (lang as SupportedLang)
    : DEFAULT_LANG;
}
