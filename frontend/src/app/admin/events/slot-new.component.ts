import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, Event, Stage, DJ } from '../../services/api.service';

@Component({
  selector: 'app-slot-new',
  standalone: true,
  imports: [FormsModule, RouterLink, TranslatePipe],
  templateUrl: './slot-new.component.html',
  styleUrl: './slot-new.component.css',
})
export class SlotNewComponent implements OnDestroy {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  eventId = this.route.snapshot.paramMap.get('id')!;

  event = signal<Event | null>(null);
  stages = signal<Stage[]>([]);
  djs = signal<DJ[]>([]);

  stage_id = '';
  dj_id = '';
  genre = '';
  slot_date = '';
  start_time = '';
  end_time = '';
  notes = '';

  minDate = computed(() => this.event()?.start_date ?? '');
  maxDate = computed(() => this.event()?.end_date ?? '');

  private subscriptions: Subscription[] = [];

  constructor() {
    this.subscriptions.push(
      this.api.getEvent(this.eventId).subscribe(e => this.event.set(e)),
      this.api.getStages(this.eventId).subscribe(s => this.stages.set(s)),
      this.api.getDJs().subscribe(d => this.djs.set(d)),
    );
  }

  submit() {
    if (!this.stage_id || !this.slot_date || !this.start_time || !this.end_time) return;

    this.api.createSlot(this.eventId, {
      stage_id: this.stage_id,
      dj_id: this.dj_id,
      genre: this.genre,
      slot_date: this.slot_date,
      start_time: this.start_time,
      end_time: this.end_time,
      notes: this.notes,
    }).subscribe(() => {
      this.router.navigate(['/admin/events', this.eventId]);
    });
  }

  ngOnDestroy() {
    this.subscriptions.forEach(s => s.unsubscribe());
  }
}
