import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-event-new',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslatePipe],
  template: `
    <div class="max-w-2xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold text-gray-900">{{ 'events.new' | translate }}</h1>
        <a routerLink="/admin/events"
           class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">
          ← {{ 'actions.back' | translate }}
        </a>
      </div>

      <form (ngSubmit)="submit()" class="bg-white border border-gray-200 rounded-lg p-6">
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'events.name' | translate }}</label>
          <input [(ngModel)]="name" name="name" type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
            required />
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'events.venue' | translate }}</label>
          <input [(ngModel)]="venue_name" name="venue_name" type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
            required />
        </div>

        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'events.startDate' | translate }}</label>
            <input [(ngModel)]="start_date" name="start_date" type="date"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              required />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ 'events.endDate' | translate }}</label>
            <input [(ngModel)]="end_date" name="end_date" type="date"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              required />
          </div>
        </div>

        <div class="flex gap-3">
          <button type="submit"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-lg transition-colors">
            {{ 'actions.create' | translate }}
          </button>
          <a routerLink="/admin/events"
             class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-lg transition-colors">
            {{ 'actions.cancel' | translate }}
          </a>
        </div>
      </form>
    </div>
  `,
})
export class EventNewComponent {
  private api = inject(ApiService);
  private router = inject(Router);
  private translate = inject(TranslateService);

  name = signal('');
  venue_name = signal('');
  start_date = signal('');
  end_date = signal('');

  submit() {
    if (!this.name() || !this.venue_name() || !this.start_date()) return;

    this.api.createEvent({
      name: this.name(),
      venue_name: this.venue_name(),
      start_date: this.start_date(),
      end_date: this.end_date(),
    }).subscribe(() => {
      this.router.navigate(['/admin/events']);
    });
  }
}
