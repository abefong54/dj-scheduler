import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, Event, Stage, Slot, DJ } from '../../services/api.service';

@Component({
  selector: 'app-event-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, TranslatePipe],
  template: `
    <div class="max-w-4xl mx-auto px-6 py-8" *ngIf="event()">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold text-gray-900">{{ event()!.name }}</h1>
        <a routerLink="/admin/events"
           class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">
          ← {{ 'actions.back' | translate }}
        </a>
      </div>

      <div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
        <p class="text-gray-600">{{ event()!.venue_name }}</p>
        <p class="text-sm text-gray-500">{{ event()!.start_date }} – {{ event()!.end_date }}</p>
      </div>

      <!-- Stages -->
      <div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-900">{{ 'stages.title' | translate }}</h2>
          <button (click)="addStage()"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
            + {{ 'stages.new' | translate }}
          </button>
        </div>
        <div class="space-y-2">
          <div *ngFor="let s of stages()" class="flex items-center gap-3 p-3 bg-gray-50 rounded">
            <span class="w-3 h-3 rounded-full" [style.background]="s.color"></span>
            <span class="text-sm font-medium text-gray-900 flex-1">{{ s.name }}</span>
            <button (click)="deleteStage(s.id)"
              class="text-red-500 hover:text-red-700 text-sm font-medium">
              {{ 'actions.delete' | translate }}
            </button>
          </div>
        </div>
      </div>

      <!-- Slots -->
      <div class="bg-white border border-gray-200 rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-900">{{ 'slots.title' | translate }}</h2>
          <a [routerLink]="['/admin/events', event()!.id, 'slots', 'new']"
             class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
            + {{ 'slots.new' | translate }}
          </a>
        </div>
        <div class="space-y-2">
          <div *ngFor="let slot of slots()" class="flex items-center gap-3 p-3 bg-gray-50 rounded">
            <span class="text-sm font-medium text-gray-900">
              {{ slot.slot_date }} {{ slot.start_time }}–{{ slot.end_time }}
            </span>
            <span class="text-sm text-gray-600 flex-1">{{ slot.dj_name || '—' }}</span>
            <button (click)="deleteSlot(slot.id)"
              class="text-red-500 hover:text-red-700 text-sm font-medium">
              {{ 'actions.delete' | translate }}
            </button>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class EventDetailComponent implements OnDestroy {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private translate = inject(TranslateService);

  event = signal<Event | null>(null);
  stages = signal<Stage[]>([]);
  slots = signal<Slot[]>([]);

  private subscriptions: Subscription[] = [];

  constructor() {
    this.loadData();
  }

  private loadData() {
    const eventId = this.route.snapshot.paramMap.get('id')!;

    const eventSub = this.api.getEvent(eventId).subscribe(e => {
      this.event.set(e);
    });

    const stagesSub = this.api.getStages(eventId).subscribe(s => {
      this.stages.set(s);
    });

    const slotsSub = this.api.getSlots(eventId).subscribe(s => {
      this.slots.set(s);
    });

    this.subscriptions.push(eventSub, stagesSub, slotsSub);
  }

  addStage() {
    const name = prompt(this.translate.instant('stages.newName'));
    if (!name) return;
    const eventId = this.event()!.id;
    this.api.createStage(eventId, { name, color: '#6366F1' }).subscribe(() => {
      this.api.getStages(eventId).subscribe(s => this.stages.set(s));
    });
  }

  deleteStage(stageId: string) {
    if (!confirm(this.translate.instant('stages.deleteConfirm'))) return;
    const eventId = this.event()!.id;
    this.api.deleteStage(eventId, stageId).subscribe(() => {
      this.api.getStages(eventId).subscribe(s => this.stages.set(s));
    });
  }

  deleteSlot(slotId: string) {
    if (!confirm(this.translate.instant('slots.deleteConfirm'))) return;
    const eventId = this.event()!.id;
    this.api.deleteSlot(eventId, slotId).subscribe(() => {
      this.api.getSlots(eventId).subscribe(s => this.slots.set(s));
    });
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
