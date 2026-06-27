import { Injectable, signal, effect } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

@Injectable({ providedIn: 'root' })
export class LanguageService {
  currentLang = signal<string>('en');

  constructor(private translate: TranslateService) {
    this.initLanguage();
    this.setupLanguageEffect();
  }

  private initLanguage() {
    const saved = localStorage.getItem('lang') || 'en';
    this.currentLang.set(saved);
  }

  private setupLanguageEffect() {
    effect(() => {
      const lang = this.currentLang();
      this.translate.use(lang);
      localStorage.setItem('lang', lang);
    });
  }

  setLanguage(lang: string) {
    this.currentLang.set(lang);
  }
}
