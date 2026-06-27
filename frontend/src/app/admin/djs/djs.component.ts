import { Component, inject, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, DJ } from '../../services/api.service';

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslatePipe],
  template: `
    <div class="max-w-4xl mx-auto px-6 py-8">
      <h1 class="text-2xl font-semibold text-gray-900 mb-6">{{ 'djs.title' | translate }}</h1>

      <!-- Add DJ form -->
      <div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">{{ 'djs.new' | translate }}</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'djs.name' | translate }}</label>
            <input [ngModel]="newDJ().name"
              (ngModelChange)="newDJ.set({name: $event})"
              name="name" type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="DJ Name" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'djs.genres' | translate }}</label>
            <input [ngModel]="newDJGenres()"
              (ngModelChange)="newDJGenres.set($event)"
              name="genres" type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="genre1, genre2, genre3" />
          </div>
          <button (click)="addDJ()"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-lg transition-colors">
            + {{ 'actions.add' | translate }}
          </button>
        </div>
      </div>

      <!-- DJ list -->
      <div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 bg-gray-50">
              <th class="text-left px-6 py-3 font-medium text-gray-700">{{ 'djs.name' | translate }}</th>
              <th class="text-left px-6 py-3 font-medium text-gray-700">{{ 'djs.genres' | translate }}</th>
              <th class="text-right px-6 py-3 font-medium text-gray-700">{{ 'actions.title' | translate }}</th>
            </tr>
          </thead>
          <tbody>
            <tr *ngFor="let dj of djs()" class="border-b border-gray-100 hover:bg-gray-50">
              <td class="px-6 py-3 font-medium text-gray-900">{{ dj.name }}</td>
              <td class="px-6 py-3 text-gray-600">
                <span *ngFor="let g of dj.genre_tags" class="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs mr-2">
                  {{ g }}
                </span>
              </td>
              <td class="px-6 py-3 text-right">
                <button (click)="deleteDJ(dj.id)"
                  class="text-red-500 hover:text-red-700 font-medium">
                  {{ 'actions.delete' | translate }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  `,
})
export class DjsComponent implements OnDestroy {
  private api = inject(ApiService);
  private translate = inject(TranslateService);

  djs = signal<DJ[]>([]);
  newDJ = signal({ name: '' });
  newDJGenres = signal('');

  private subscriptions: Subscription[] = [];

  constructor() {
    this.loadDJs();
  }

  private loadDJs() {
    const sub = this.api.getDJs().subscribe(djs => {
      this.djs.set(djs);
    });
    this.subscriptions.push(sub);
  }

  addDJ() {
    const name = this.newDJ().name;
    const genres = this.newDJGenres()
      .split(',')
      .map(g => g.trim())
      .filter(g => g.length > 0);

    if (!name || genres.length === 0) {
      alert(this.translate.instant('djs.fillRequired'));
      return;
    }

    this.api.createDJ({ name, genre_tags: genres }).subscribe(() => {
      this.newDJ.set({ name: '' });
      this.newDJGenres.set('');
      this.loadDJs();
    });
  }

  deleteDJ(id: string) {
    if (!confirm(this.translate.instant('djs.deleteConfirm'))) return;
    this.api.deleteDJ(id).subscribe(() => {
      this.loadDJs();
    });
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
