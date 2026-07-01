import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, Event, Stage, Slot, DJ } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';
import { ScheduleExportService } from '../../services/schedule-export.service';
import { slotDurationMins } from '../../shared/slot-time.util';
import { parseLocalDate } from '../../shared/date.util';
import { AdminShellComponent } from '../../shared/admin-shell.component';
import { StatusBadgeComponent } from '../../shared/status-badge.component';
import { ButtonComponent } from '../../shared/button.component';
import { TableSort } from '../../shared/table-sort';

@Component({
  selector: 'app-event-detail',
  standalone: true,
  imports: [FormsModule, RouterLink, TranslatePipe, AdminShellComponent, StatusBadgeComponent, ButtonComponent],
  templateUrl: './event-detail.component.html',
  styleUrl: './event-detail.component.css',
})
export class EventDetailComponent implements OnDestroy {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private exporter = inject(ScheduleExportService);

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

  // ── Slot search + sort (shared TableSort engine, EL-028) ────────────
  slotTable = new TableSort<Slot>(() => this.slotsForSelectedDate(), {
    searchKeys: () => ['dj_name', 'stage_name', 'genre'],
    initialSortKey: 'start_time',
  });

  // ── Inline add row ──────────────────────────────────────────────────
  addRowActive = signal(false);
  addStageId = '';
  addDjId = '';
  addGenre = '';
  addDate = '';
  addStart = '';
  addDuration = 60;
  addNotes = '';
  addConflictError = signal<string | null>(null);

  defaultStageId = computed(() => this.stages().length === 1 ? this.stages()[0].id : '');

  defaultStartTime = computed(() => {
    const slots = this.slotsForSelectedDate();
    return slots.length > 0 ? slots[slots.length - 1].end_time : '18:00';
  });

  // ── Genre ↔ DJ bidirectional filtering (add row) ──────────────────
  private allGenres(): string[] {
    const all = new Set<string>();
    this.djs().forEach(d => d.genre_tags.forEach(g => all.add(g)));
    return [...all].sort();
  }

  genreOptionsForAdd(): string[] {
    return this.addDjId
      ? (this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [])
      : this.allGenres();
  }

  filteredDjsForAdd(): DJ[] {
    return this.addGenre
      ? this.djs().filter(d => d.genre_tags.includes(this.addGenre))
      : this.djs();
  }

  // ── EL-042: certification-aware DJ options ────────────────────────
  // A student DJ is cleared for a genre only if they hold that certification
  // (case-insensitive). Graduates (is_student === false) bypass the gate, and
  // with no target genre there's nothing to gate.
  isCertifiedFor(dj: DJ, genre: string): boolean {
    if (dj.is_student === false) return true;
    if (!genre) return true;
    const g = genre.toLowerCase();
    return (dj.certifications ?? []).some(c => c.toLowerCase() === g);
  }

  // Certified DJs first; uncertified are flagged (not hidden) so the teacher can
  // still override.
  djOptionsForAdd(): { dj: DJ; certified: boolean }[] {
    return this.filteredDjsForAdd()
      .map(dj => ({ dj, certified: this.isCertifiedFor(dj, this.addGenre) }))
      .sort((a, b) => Number(b.certified) - Number(a.certified));
  }

  addCertWarning(): string | null {
    const dj = this.djs().find(d => d.id === this.addDjId);
    if (!dj || this.isCertifiedFor(dj, this.addGenre)) return null;
    return this.translate.instant('slots.cert.warning', { djName: dj.name, genre: this.addGenre });
  }

  onAddDjChange() {
    const djGenres = this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
    if (this.addGenre && this.addDjId && !djGenres.includes(this.addGenre)) {
      this.addGenre = '';
    }
  }

  onAddGenreChange() {
    const djGenres = this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
    if (this.addDjId && this.addGenre && !djGenres.includes(this.addGenre)) {
      this.addDjId = '';
    }
  }

  activateAddRow() {
    this.addStageId = this.defaultStageId();
    this.addDjId = '';
    this.addGenre = '';
    this.addDate = this.selectedDate();
    this.addStart = this.defaultStartTime();
    this.addDuration = 60;
    this.addNotes = '';
    this.addConflictError.set(null);
    this.addRowActive.set(true);
  }

  cancelAddRow() {
    this.addConflictError.set(null);
    this.addRowActive.set(false);
  }

  submitAddRow() {
    if (!this.addStageId || !this.addDate || !this.addStart) return;
    const eventId = this.event()!.id;
    this.api.createSlot(eventId, {
      stage_id: this.addStageId,
      dj_id: this.addDjId,
      genre: this.addGenre,
      slot_date: this.addDate,
      start_time: this.addStart,
      end_time: this.addMinutes(this.addStart, this.addDuration),
      notes: this.addNotes,
    }).subscribe({
      next: () => {
        this.api.getSlots(eventId).subscribe(s => {
          this.slots.set(s);
          this.activateAddRow(); // reset with fresh pre-fills (also clears the error) for the next slot
        });
      },
      error: (err) => this.addConflictError.set(this.conflictMessage(err, this.addDjId, this.addStageId)),
    });
  }

  // Maps a slot write error to a user-facing message: a 409 reports the specific
  // DJ double-booking or stage overlap; anything else is a generic failure.
  private conflictMessage(err: unknown, djId: string, stageId: string): string {
    const e = err as { status?: number; error?: { type?: string } };
    if (e?.status === 409) {
      if (e.error?.type === 'dj_double_booked') {
        const djName = this.djs().find(d => d.id === djId)?.name ?? '';
        return this.translate.instant('slots.error.djDoubleBooked', { djName });
      }
      if (e.error?.type === 'stage_overlap') {
        const stageName = this.stages().find(s => s.id === stageId)?.name ?? '';
        return this.translate.instant('slots.error.stageOverlap', { stageName });
      }
    }
    return this.translate.instant('slots.error.generic');
  }

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
    const mins = slotDurationMins(start, end);
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
    return parseLocalDate(d);
  }

  viewPublic() {
    window.open(`/events/${this.event()!.id}`, '_blank');
  }

  shareViaLine() {
    const url = encodeURIComponent(`${window.location.origin}/events/${this.event()!.id}`);
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }

  exportXlsx() {
    const e = this.event();
    if (!e) return;
    this.exporter.download(e.name, this.slots(), {
      date: this.translate.instant('export.date'),
      stage: this.translate.instant('export.stage'),
      timeSlot: this.translate.instant('export.timeSlot'),
      dj: this.translate.instant('export.dj'),
      genre: this.translate.instant('export.genre'),
      notes: this.translate.instant('export.notes'),
    });
  }

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

  // ── Genre ↔ DJ bidirectional filtering (edit row) ─────────────────
  genreOptionsForEdit(): string[] {
    return this.editSlotDjId
      ? (this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [])
      : this.allGenres();
  }

  filteredDjsForEdit(): DJ[] {
    return this.editSlotGenre
      ? this.djs().filter(d => d.genre_tags.includes(this.editSlotGenre))
      : this.djs();
  }

  // EL-042: certification-aware options for the inline edit DJ dropdown.
  djOptionsForEdit(): { dj: DJ; certified: boolean }[] {
    return this.filteredDjsForEdit()
      .map(dj => ({ dj, certified: this.isCertifiedFor(dj, this.editSlotGenre) }))
      .sort((a, b) => Number(b.certified) - Number(a.certified));
  }

  editCertWarning(): string | null {
    const dj = this.djs().find(d => d.id === this.editSlotDjId);
    if (!dj || this.isCertifiedFor(dj, this.editSlotGenre)) return null;
    return this.translate.instant('slots.cert.warning', { djName: dj.name, genre: this.editSlotGenre });
  }

  onEditDjChange() {
    const djGenres = this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
    if (this.editSlotGenre && this.editSlotDjId && !djGenres.includes(this.editSlotGenre)) {
      this.editSlotGenre = '';
    }
  }

  onEditGenreChange() {
    const djGenres = this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
    if (this.editSlotDjId && this.editSlotGenre && !djGenres.includes(this.editSlotGenre)) {
      this.editSlotDjId = '';
    }
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

  private addMinutes(time: string, mins: number): string {
    const total = this.toMins(time) + mins;
    const h = Math.floor(total / 60) % 24;
    const m = total % 60;
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
  }

  editingSlotId = signal<string | null>(null);
  editSlotStageId = '';
  editSlotDjId = '';
  editSlotGenre = '';
  editSlotDate = '';
  editSlotStart = '';
  editSlotDuration = 60;
  editSlotNotes = '';
  editConflictError = signal<string | null>(null);

  startEdit(slot: Slot) {
    this.editSlotStageId = slot.stage_id;
    this.editSlotDjId = slot.dj_id;
    this.editSlotGenre = slot.genre;
    this.editSlotDate = slot.slot_date;
    this.editSlotStart = slot.start_time;
    this.editSlotDuration = slotDurationMins(slot.start_time, slot.end_time);
    this.editSlotNotes = slot.notes;
    this.editConflictError.set(null);
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
    }).subscribe({
      next: () => {
        this.api.getSlots(eventId).subscribe(s => this.slots.set(s));
        this.editConflictError.set(null);
        this.editingSlotId.set(null);
      },
      error: (err) => this.editConflictError.set(this.conflictMessage(err, this.editSlotDjId, this.editSlotStageId)),
    });
  }

  cancelEdit() {
    this.editConflictError.set(null);
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
