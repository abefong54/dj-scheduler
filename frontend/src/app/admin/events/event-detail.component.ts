import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot, DJ } from '../../services/api.service';

@Component({
  selector: 'app-event-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslatePipe],
  template: `
    <div class="max-w-6xl mx-auto px-6 py-8" *ngIf="event">
      <!-- Header -->
      <div class="mb-6">
        <a routerLink="/admin/events" class="text-sm text-indigo-600 hover:underline mb-2 inline-block">
          ← Events
        </a>
        <div class="flex items-start justify-between flex-wrap gap-3">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900">{{ event.name }}</h1>
            <p class="text-sm text-gray-500 mt-1">{{ event.venue_name }} · {{ formatDates(event) }}</p>
          </div>
          <div class="flex gap-3">
            <button (click)="shareViaLine()"
              class="bg-green-500 hover:bg-green-600 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventDetail.shareViaLine' | translate }}
            </button>
            <a [href]="publicUrl" target="_blank"
               class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventDetail.viewPublic' | translate }} →
            </a>
          </div>
        </div>
      </div>

      <!-- Two-column layout -->
      <div class="flex gap-6 flex-col md:flex-row">

        <!-- Stages panel -->
        <div class="w-full md:w-56 shrink-0">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
            {{ 'stages.title' | translate }}
          </p>
          <ul class="space-y-2 mb-4">
            <li *ngFor="let s of stages"
                class="flex items-center gap-2 bg-white border border-gray-200 rounded-lg px-3 py-2">
              <span class="w-3 h-3 rounded-full shrink-0" [style.background]="s.color"></span>
              <span class="text-sm text-gray-800 flex-1 truncate">{{ s.name }}</span>
              <button (click)="deleteStage(s.id)"
                class="text-gray-400 hover:text-red-500 transition-colors text-sm ml-auto"
                aria-label="Delete stage">✕</button>
            </li>
          </ul>
          <form (ngSubmit)="addStage()" class="space-y-2">
            <input [(ngModel)]="newStage.name" name="stage_name"
              [placeholder]="'stages.name' | translate"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            <div class="flex gap-2 items-center">
              <input [(ngModel)]="newStage.color" name="stage_color" type="color"
                class="h-9 w-9 rounded border border-gray-300 cursor-pointer" />
              <button type="submit"
                class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-md text-sm transition-colors">
                {{ 'stages.add' | translate }}
              </button>
            </div>
          </form>
        </div>

        <!-- Schedule panel -->
        <div class="flex-1 min-w-0">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
            {{ 'schedule.title' | translate }}
          </p>

          <!-- Add slot form -->
          <form (ngSubmit)="addSlot()" class="bg-white border border-gray-200 rounded-lg p-3 mb-4">
            <div class="flex flex-wrap gap-2">
              <select [(ngModel)]="newSlot.stage_id" name="slot_stage"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">{{ 'slots.stage' | translate }}</option>
                <option *ngFor="let s of stages" [value]="s.id">{{ s.name }}</option>
              </select>
              <input [(ngModel)]="newSlot.slot_date" name="slot_date" type="date"
                [min]="event.start_date" [max]="event.end_date"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <input [(ngModel)]="newSlot.start_time" name="slot_start" type="time"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <input [(ngModel)]="newSlot.end_time" name="slot_end" type="time"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <select [(ngModel)]="newSlot.dj_id" name="slot_dj"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">{{ 'slots.unassigned' | translate }}</option>
                <option *ngFor="let d of djs" [value]="d.id">{{ d.name }}</option>
              </select>
              <input [(ngModel)]="newSlot.notes" name="slot_notes"
                [placeholder]="'slots.notes' | translate"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-24" />
              <button type="submit"
                class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-3 py-2 rounded-md text-sm transition-colors"
                aria-label="Add slot">+</button>
            </div>
          </form>

          <!-- Slot list -->
          <div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <p *ngIf="slots.length===0"
               class="text-gray-400 text-sm text-center py-8">
              {{ 'slots.noSlots' | translate }}
            </p>
            <table *ngIf="slots.length>0" class="w-full text-sm">
              <thead class="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.date' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.stage' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.start' | translate }}–{{ 'slots.end' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.dj' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.notes' | translate }}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr *ngFor="let s of slots" class="hover:bg-gray-50">
                  <td class="px-4 py-2 text-gray-700">{{ s.slot_date }}</td>
                  <td class="px-4 py-2">
                    <div class="flex items-center gap-1.5">
                      <span class="w-2.5 h-2.5 rounded-full" [style.background]="stageColor(s.stage_id)"></span>
                      <span class="text-gray-700">{{ s.stage_name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-2 text-gray-700">{{ s.start_time }}–{{ s.end_time }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ s.dj_name || '—' }}</td>
                  <td class="px-4 py-2 text-gray-500">{{ s.notes }}</td>
                  <td class="px-4 py-2 text-right">
                    <button (click)="deleteSlot(s.id)"
                      class="text-gray-400 hover:text-red-500 transition-colors"
                      aria-label="Delete slot">✕</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class EventDetailComponent implements OnInit {
  event: Event | null = null;
  stages: Stage[] = [];
  slots: Slot[] = [];
  djs: DJ[] = [];
  newStage = { name: '', color: '#6366F1' };
  newSlot = { stage_id: '', dj_id: '', slot_date: '', start_time: '', end_time: '', notes: '' };

  private eventId = '';

  constructor(private api: ApiService, private route: ActivatedRoute) {}

  ngOnInit() {
    this.eventId = this.route.snapshot.paramMap.get('id')!;
    this.api.getEvent(this.eventId).subscribe(e => this.event = e);
    this.api.getDJs().subscribe(d => this.djs = d);
    this.loadStages();
    this.loadSlots();
  }

  loadStages() { this.api.getStages(this.eventId).subscribe(s => this.stages = s); }
  loadSlots() { this.api.getSlots(this.eventId).subscribe(s => this.slots = s); }

  stageColor(stageId: string): string {
    return this.stages.find(s => s.id === stageId)?.color ?? '#6366F1';
  }

  formatDates(e: Event): string {
    return e.start_date === e.end_date ? e.start_date : `${e.start_date} – ${e.end_date}`;
  }

  addStage() {
    if (!this.newStage.name.trim()) return;
    this.api.createStage(this.eventId, this.newStage).subscribe(() => {
      this.newStage = { name: '', color: '#6366F1' };
      this.loadStages();
    });
  }

  deleteStage(stageId: string) {
    this.api.deleteStage(this.eventId, stageId).subscribe(() => this.loadStages());
  }

  addSlot() {
    if (!this.newSlot.stage_id || !this.newSlot.slot_date || !this.newSlot.start_time || !this.newSlot.end_time) return;
    this.api.createSlot(this.eventId, this.newSlot).subscribe(() => {
      this.newSlot = { stage_id: '', dj_id: '', slot_date: '', start_time: '', end_time: '', notes: '' };
      this.loadSlots();
    });
  }

  deleteSlot(slotId: string) {
    this.api.deleteSlot(this.eventId, slotId).subscribe(() => this.loadSlots());
  }

  get publicUrl(): string {
    return `${window.location.origin}/events/${this.eventId}`;
  }

  shareViaLine() {
    window.open(`https://line.me/R/msg/text/?${encodeURIComponent(this.publicUrl)}`, '_blank');
  }
}
