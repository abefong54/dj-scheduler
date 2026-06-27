import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, DJ } from '../../services/api.service';

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslatePipe],
  template: `
    <div class="max-w-2xl mx-auto px-6 py-8">
      <h1 class="text-2xl font-semibold text-gray-900 mb-6">{{ 'djs.title' | translate }}</h1>

      <!-- Add DJ form -->
      <div class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm mb-6">
        <form (ngSubmit)="add()" class="flex flex-wrap gap-3 items-end">
          <div class="flex-1 min-w-32">
            <label class="block text-xs font-medium text-gray-600 mb-1">{{ 'djs.name' | translate }} *</label>
            <input [(ngModel)]="newName" name="name" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <div class="flex-1 min-w-40">
            <label class="block text-xs font-medium text-gray-600 mb-1">{{ 'djs.genre' | translate }}</label>
            <input [(ngModel)]="newGenre" name="genre"
              [attr.placeholder]="'djs.genrePlaceholder' | translate"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <button type="submit"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors h-[38px]">
            {{ 'djs.add' | translate }}
          </button>
        </form>
      </div>

      <!-- DJ list -->
      <p *ngIf="djs.length===0" class="text-gray-400 text-sm text-center py-8">
        {{ 'djs.noDJs' | translate }}
      </p>
      <ul *ngIf="djs.length>0" class="bg-white border border-gray-200 rounded-lg divide-y divide-gray-100 shadow-sm">
        <li *ngFor="let d of djs"
            class="flex items-center gap-3 px-4 py-3">
          <div class="flex-1 min-w-0">
            <span class="font-medium text-gray-900 text-sm">{{ d.name }}</span>
            <div *ngIf="d.genre_tags?.length" class="flex flex-wrap gap-1.5 mt-1">
              <span *ngFor="let tag of d.genre_tags"
                class="bg-indigo-50 text-indigo-700 text-xs font-medium px-2.5 py-0.5 rounded-full">
                {{ tag }}
              </span>
            </div>
          </div>
          <button (click)="delete(d.id)"
            class="text-gray-400 hover:text-red-500 transition-colors shrink-0"
            [attr.aria-label]="'actions.delete' | translate">✕</button>
        </li>
      </ul>
    </div>
  `,
})
export class DjsComponent implements OnInit {
  djs: DJ[] = [];
  newName = '';
  newGenre = '';

  constructor(private api: ApiService) {}

  ngOnInit() { this.load(); }

  load() { this.api.getDJs().subscribe(d => this.djs = d); }

  add() {
    if (!this.newName.trim()) return;
    const genre_tags = this.newGenre
      .split(',')
      .map(t => t.trim())
      .filter(t => t.length > 0);
    this.api.createDJ({ name: this.newName.trim(), genre_tags }).subscribe(() => {
      this.newName = '';
      this.newGenre = '';
      this.load();
    });
  }

  delete(id: string) {
    this.api.deleteDJ(id).subscribe(() => this.load());
  }
}
