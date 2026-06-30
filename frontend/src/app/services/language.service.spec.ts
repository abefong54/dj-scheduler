import { TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';

import { LanguageService } from './language.service';

describe('LanguageService', () => {
  function make(): LanguageService {
    TestBed.configureTestingModule({
      providers: [provideTranslateService()],
    });
    return TestBed.inject(LanguageService);
  }

  afterEach(() => localStorage.clear());

  it('restores a supported saved language', () => {
    localStorage.setItem('lang', 'zh-TW');
    expect(make().currentLang()).toBe('zh-TW');
  });

  it('falls back to English for a legacy / unsupported saved language', () => {
    // A stale "zh" (from an earlier build) would 404 its translation file and
    // leave the whole UI showing raw keys — coerce it back to a supported lang.
    localStorage.setItem('lang', 'zh');
    expect(make().currentLang()).toBe('en');
  });

  it('falls back to English for a garbage saved language', () => {
    localStorage.setItem('lang', 'null');
    expect(make().currentLang()).toBe('en');
  });

  it('defaults to English when nothing is saved', () => {
    expect(make().currentLang()).toBe('en');
  });

  it('ignores an unsupported language passed to setLanguage', () => {
    const svc = make();
    svc.setLanguage('zh');
    expect(svc.currentLang()).toBe('en');
  });

  it('accepts a supported language passed to setLanguage', () => {
    const svc = make();
    svc.setLanguage('zh-TW');
    expect(svc.currentLang()).toBe('zh-TW');
  });

  it('heals a stale saved language in localStorage', () => {
    localStorage.setItem('lang', 'zh');
    make();
    TestBed.tick();
    expect(localStorage.getItem('lang')).toBe('en');
  });
});
