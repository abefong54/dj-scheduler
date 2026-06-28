import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, Event, Stage, Slot, DJ } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';

@Component({
  selector: 'app-event-detail',
  standalone: true,
  imports: [FormsModule, RouterLink, TranslatePipe],
  templateUrl: './event-detail.component.html',
  styleUrl: './event-detail.component.css',
})
export class EventDetailComponent implements OnDestroy {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);

  event = signal<Event | null>(null);
  stages = signal<Stage[]>([]);
  slots = signal<Slot[]>([]);
  djs = signal<DJ[]>([]);

  selectedDate = signal('');

  dates = computed(() => {
    const slotDates = [...new Set(this.slots().map(s => s.slot_date))].sort();
    const start = this.event()?.start_date;
    if (start && !slotDates.includes(start)) slotDates.unshift(start);
    return slotDates.sort();
  });

  slotsForSelectedDate = computed(() =>
    this.slots()
      .filter(s => s.slot_date === this.selectedDate())
      .sort((a, b) => a.start_time.localeCompare(b.start_time))
  );

  tabDate(d: string): string {
    const [, month, day] = d.split('-');
    return `${parseInt(month)}/${parseInt(day)}`;
  }

  private subscriptions: Subscription[] = [];

  constructor() {
    this.loadData();
  }

  private loadData() {
    const eventId = this.route.snapshot.paramMap.get('id')!;

    this.subscriptions.push(
      this.api.getEvent(eventId).subscribe(e => {
        this.event.set(e);
        if (this.selectedDate() === '') this.selectedDate.set(e.start_date);
      }),
      this.api.getStages(eventId).subscribe(s => this.stages.set(s)),
      this.api.getSlots(eventId).subscribe(s => this.slots.set(s)),
      this.api.getDJs().subscribe(d => this.djs.set(d)),
    );
  }

  private stageMap = computed(() =>
    new Map(this.stages().map(s => [s.id, s]))
  );

  slotsByDate = computed(() => {
    const map = new Map<string, Slot[]>();
    for (const slot of this.slots()) {
      const arr = map.get(slot.slot_date) ?? [];
      arr.push(slot);
      map.set(slot.slot_date, arr);
    }
    return [...map.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([date, slots]) => ({
        date,
        slots: [...slots].sort((a, b) => a.start_time.localeCompare(b.start_time)),
      }));
  });

  dateRange(): string {
    const e = this.event();
    if (!e) return '';
    const start = this.parseDate(e.start_date);
    const end = this.parseDate(e.end_date);
    const opts: Intl.DateTimeFormatOptions = { month: 'long', day: 'numeric' };
    if (e.start_date === e.end_date) {
      return start.toLocaleDateString('en-US', { ...opts, year: 'numeric' });
    }
    if (start.getMonth() === end.getMonth()) {
      return `${start.toLocaleDateString('en-US', opts)}–${end.getDate()}, ${start.getFullYear()}`;
    }
    return `${start.toLocaleDateString('en-US', opts)} – ${end.toLocaleDateString('en-US', { ...opts, year: 'numeric' })}`;
  }

  formatDate(d: string): string {
    return this.parseDate(d).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
  }

  stageColor(stageId: string): string {
    return this.stageMap().get(stageId)?.color ?? '#9CA3AF';
  }

  duration(start: string, end: string): string {
    const mins = this.toMins(end) - this.toMins(start);
    if (mins < 60) return `${mins}m`;
    const h = Math.floor(mins / 60);
    const m = mins % 60;
    return m === 0 ? `${h}h` : `${h}h ${m}m`;
  }

  private toMins(t: string): number {
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  private parseDate(d: string): Date {
    return new Date(d + 'T00:00:00');
  }

  viewPublic() {
    window.open(`/events/${this.event()!.id}`, '_blank');
  }

  shareViaLine() {
    const url = encodeURIComponent(`${window.location.origin}/events/${this.event()!.id}`);
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }

  exportXlsx() { /* implemented in Task 6 */ }

  showStageModal = signal(false);
  newStageName = '';
  newStageColor = '#06b6d4';
  readonly stageColors = [
    '#06b6d4', // cyan-500
    '#e879f9', // fuchsia-400
    '#84cc16', // lime-500
    '#f97316', // orange-500
    '#0ea5e9', // sky-500
    '#ec4899', // pink-500
    '#facc15', // yellow-400
    '#14b8a6', // teal-500
  ];

  addStage() {
    this.newStageName = '';
    this.newStageColor = '#06b6d4';
    this.showStageModal.set(true);
  }

  submitNewStage() {
    if (!this.newStageName.trim()) return;
    const eventId = this.event()!.id;
    this.api.createStage(eventId, { name: this.newStageName.trim(), color: this.newStageColor }).subscribe(() => {
      this.api.getStages(eventId).subscribe(s => this.stages.set(s));
      this.showStageModal.set(false);
    });
  }

  cancelStage() {
    this.showStageModal.set(false);
  }

  djGenresForSlot(): string[] {
    return this.djs().find(d => d.id === this.newSlotDjId)?.genre_tags ?? [];
  }

  djGenresForEdit(): string[] {
    return this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
  }

  readonly durationOptions = [
    { mins: 30,  label: '30m' },
    { mins: 45,  label: '45m' },
    { mins: 60,  label: '1h' },
    { mins: 90,  label: '1h 30m' },
    { mins: 120, label: '2h' },
    { mins: 150, label: '2h 30m' },
    { mins: 180, label: '3h' },
  ];

  showSlotModal = signal(false);
  newSlotStageId = '';
  newSlotDjId = '';
  newSlotGenre = '';
  newSlotDate = '';
  newSlotStart = '';
  newSlotDuration = 60;
  newSlotNotes = '';

  addSlot() {
    this.newSlotStageId = '';
    this.newSlotDjId = '';
    this.newSlotGenre = '';
    this.newSlotDate = '';
    this.newSlotStart = '';
    this.newSlotDuration = 60;
    this.newSlotNotes = '';
    this.showSlotModal.set(true);
  }

  submitNewSlot() {
    if (!this.newSlotStageId || !this.newSlotDate || !this.newSlotStart) return;
    const eventId = this.event()!.id;
    this.api.createSlot(eventId, {
      stage_id: this.newSlotStageId,
      dj_id: this.newSlotDjId,
      genre: this.newSlotGenre,
      slot_date: this.newSlotDate,
      start_time: this.newSlotStart,
      end_time: this.addMinutes(this.newSlotStart, this.newSlotDuration),
      notes: this.newSlotNotes,
    }).subscribe(() => {
      this.api.getSlots(eventId).subscribe(s => this.slots.set(s));
      this.showSlotModal.set(false);
    });
  }

  private addMinutes(time: string, mins: number): string {
    const total = this.toMins(time) + mins;
    const h = Math.floor(total / 60) % 24;
    const m = total % 60;
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
  }

  cancelSlot() {
    this.showSlotModal.set(false);
  }

  editingSlotId = signal<string | null>(null);
  editSlotStageId = '';
  editSlotDjId = '';
  editSlotGenre = '';
  editSlotDate = '';
  editSlotStart = '';
  editSlotDuration = 60;
  editSlotNotes = '';

  startEdit(slot: Slot) {
    this.editSlotStageId = slot.stage_id;
    this.editSlotDjId = slot.dj_id;
    this.editSlotGenre = slot.genre;
    this.editSlotDate = slot.slot_date;
    this.editSlotStart = slot.start_time;
    this.editSlotDuration = this.toMins(slot.end_time) - this.toMins(slot.start_time);
    this.editSlotNotes = slot.notes;
    this.editingSlotId.set(slot.id);
  }

  saveEdit(slotId: string) {
    const eventId = this.event()!.id;
    this.api.updateSlot(eventId, slotId, {
      stage_id: this.editSlotStageId,
      dj_id: this.editSlotDjId,
      genre: this.editSlotGenre,
      slot_date: this.editSlotDate,
      start_time: this.editSlotStart,
      end_time: this.addMinutes(this.editSlotStart, Number(this.editSlotDuration)),
      notes: this.editSlotNotes,
    }).subscribe(() => {
      this.api.getSlots(eventId).subscribe(s => this.slots.set(s));
      this.editingSlotId.set(null);
    });
  }

  cancelEdit() {
    this.editingSlotId.set(null);
  }

  async deleteStage(stageId: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('stages.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    const eventId = this.event()!.id;
    this.api.deleteStage(eventId, stageId).subscribe(() => {
      this.api.getStages(eventId).subscribe(s => this.stages.set(s));
    });
  }

  async deleteSlot(slotId: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('slots.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    const eventId = this.event()!.id;
    this.api.deleteSlot(eventId, slotId).subscribe(() => {
      this.api.getSlots(eventId).subscribe(s => this.slots.set(s));
    });
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
