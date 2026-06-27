import { Component, inject } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';
import { LanguageService } from './services/language.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, TranslatePipe, CommonModule],
  template: `
    <nav class="h-14 bg-slate-900 flex items-center justify-between px-6 sticky top-0 z-50">
      <div class="flex items-center gap-8">
        <span class="text-white font-semibold text-lg">EventLineup</span>
        <a routerLink="/admin/events" routerLinkActive="border-b-2 border-indigo-400 text-white"
           class="text-slate-300 hover:text-white pb-1 transition-colors text-sm font-medium">
          {{ 'nav.events' | translate }}
        </a>
        <a routerLink="/admin/djs" routerLinkActive="border-b-2 border-indigo-400 text-white"
           class="text-slate-300 hover:text-white pb-1 transition-colors text-sm font-medium">
          {{ 'nav.djs' | translate }}
        </a>
      </div>
      <div class="flex gap-1">
        <button (click)="langService.setLanguage('en')"
          class="px-3 py-1 rounded text-xs font-medium transition-colors"
          [class]="langService.currentLang() === 'en' ? 'bg-white text-slate-900' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'">
          EN
        </button>
        <button (click)="langService.setLanguage('zh-TW')"
          class="px-3 py-1 rounded text-xs font-medium transition-colors"
          [class]="langService.currentLang() === 'zh-TW' ? 'bg-white text-slate-900' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'">
          中文
        </button>
      </div>
    </nav>
    <main class="min-h-[calc(100dvh-3.5rem)] bg-gray-50">
      <router-outlet />
    </main>
  `,
})
export class AppComponent {
  langService = inject(LanguageService);
}
