import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-event-new',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslatePipe],
  template: `
    <div class="max-w-lg mx-auto px-6 py-8">
      <a routerLink="/admin/events" class="text-sm text-indigo-600 hover:underline mb-6 inline-block">
        ← Back to Events
      </a>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
        <h1 class="text-xl font-semibold text-gray-900 mb-6">{{ 'eventNew.title' | translate }}</h1>
        <form (ngSubmit)="submit()" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.name' | translate }} *
            </label>
            <input [(ngModel)]="form.name" name="name" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.venue' | translate }} *
            </label>
            <input [(ngModel)]="form.venue_name" name="venue" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.startDate' | translate }} *
            </label>
            <input [(ngModel)]="form.start_date" name="start_date" type="date" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.endDate' | translate }}
            </label>
            <input [(ngModel)]="form.end_date" name="end_date" type="date"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div class="flex gap-3 justify-end pt-2">
            <a routerLink="/admin/events"
               class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'actions.cancel' | translate }}
            </a>
            <button type="submit"
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventNew.create' | translate }} →
            </button>
          </div>
        </form>
      </div>
    </div>
  `,
})
export class EventNewComponent {
  form = { name: '', venue_name: '', start_date: '', end_date: '' };

  constructor(private api: ApiService, private router: Router) {}

  submit() {
    if (!this.form.name || !this.form.venue_name || !this.form.start_date) return;
    const payload = {
      ...this.form,
      end_date: this.form.end_date || this.form.start_date,
    };
    this.api.createEvent(payload).subscribe(e => {
      this.router.navigate(['/admin/events', e.id]);
    });
  }
}
