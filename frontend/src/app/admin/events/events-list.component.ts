import { Component, computed, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, Event } from '../../services/api.service';

@Component({
  selector: 'app-events-list',
  standalone: true,
  imports: [CommonModule, RouterLink, TranslatePipe],
  template: `
    <div class="max-w-4xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold text-gray-900">{{ 'events.title' | translate }}</h1>
        <a routerLink="/admin/events/new"
           class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
          + {{ 'events.new' | translate }}
        </a>
      </div>

      <!-- Tabs -->
      <div class="flex gap-4 border-b border-gray-200 mb-6">
        <button (click)="activeTab.set('upcoming')"
          class="pb-3 text-sm font-medium transition-colors"
          [class]="activeTab() === 'upcoming'
            ? 'border-b-2 border-indigo-600 text-indigo-600'
            : 'text-gray-500 hover:text-gray-700'">
          {{ 'events.upcoming' | translate }}
        </button>
        <button (click)="activeTab.set('past')"
          class="pb-3 text-sm font-medium transition-colors"
          [class]="activeTab() === 'past'
            ? 'border-b-2 border-indigo-600 text-indigo-600'
            : 'text-gray-500 hover:text-gray-700'">
          {{ 'events.past' | translate }}
        </button>
      </div>

      <!-- Event cards -->
      <div class="space-y-3">
        <ng-container *ngIf="activeTab() === 'upcoming'">
          <p *ngIf="upcoming().length === 0" class="text-gray-500 text-sm py-8 text-center">
            {{ 'events.noUpcoming' | translate }}
          </p>
          <div *ngFor="let e of upcoming()"
               class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900">{{ e.name }}</p>
                <p class="text-sm text-gray-500 mt-0.5">{{ e.venue_name }} · {{ formatDates(e) }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3 mt-3">
              <a [routerLink]="['/admin/events', e.id]"
                 class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.view' | translate }}
              </a>
              <button (click)="duplicate(e.id)"
                class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.duplicate' | translate }}
              </button>
              <button (click)="delete(e.id)"
                class="text-red-500 hover:text-red-700 text-sm font-medium transition-colors ml-auto">
                {{ 'actions.delete' | translate }}
              </button>
            </div>
          </div>
        </ng-container>

        <ng-container *ngIf="activeTab() === 'past'">
          <p *ngIf="past().length === 0" class="text-gray-500 text-sm py-8 text-center">
            {{ 'events.noPast' | translate }}
          </p>
          <div *ngFor="let e of past()"
               class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm opacity-75">
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900">{{ e.name }}</p>
                <p class="text-sm text-gray-500 mt-0.5">{{ e.venue_name }} · {{ formatDates(e) }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3 mt-3">
              <a [routerLink]="['/admin/events', e.id]"
                 class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.view' | translate }}
              </a>
              <button (click)="duplicate(e.id)"
                class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.duplicate' | translate }}
              </button>
            </div>
          </div>
        </ng-container>
      </div>
    </div>
  `,
})
export class EventsListComponent {
  private api = inject(ApiService);
  private translate = inject(TranslateService);

  activeTab = signal<'upcoming' | 'past'>('upcoming');
  today = new Date().toISOString().slice(0, 10);

  events = signal<Event[]>([]);

  upcoming = computed(() => {
    const evts = this.events();
    return evts.filter(e => e.end_date >= this.today);
  });

  past = computed(() => {
    const evts = this.events();
    return evts.filter(e => e.end_date < this.today);
  });

  constructor() {
    this.loadEvents();
  }

  private loadEvents() {
    this.api.getEvents().subscribe(events => {
      this.events.set(events);
    });
  }

  formatDates(e: Event): string {
    if (e.start_date === e.end_date) return e.start_date;
    return `${e.start_date} – ${e.end_date}`;
  }

  delete(id: string) {
    if (!confirm(this.translate.instant('events.deleteConfirm'))) return;
    this.api.deleteEvent(id).subscribe(() => this.loadEvents());
  }

  duplicate(id: string) {
    this.api.duplicateEvent(id).subscribe(() => this.loadEvents());
  }
}
