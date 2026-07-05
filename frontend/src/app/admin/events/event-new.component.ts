import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService } from '../../services/api.service';
import { EventsStore } from '../../services/events-store';
import { AdminShellComponent } from '../../shared/admin-shell.component';
import { SubmitGuard } from '../../shared/submit-guard';

@Component({
  selector: 'app-event-new',
  standalone: true,
  imports: [FormsModule, RouterLink, TranslatePipe, AdminShellComponent],
  templateUrl: './event-new.component.html',
  styleUrl: './event-new.component.css',
})
export class EventNewComponent {
  private api = inject(ApiService);
  private router = inject(Router);
  private translate = inject(TranslateService);
  private store = inject(EventsStore);

  name = signal('');
  venue_name = signal('');
  start_date = signal('');
  end_date = signal('');
  genres_input = '';
  readonly submitting = new SubmitGuard();

  // Flips true on the first submit attempt so the template can reveal per-field
  // errors. Before that we stay quiet — no red text on a pristine form.
  readonly submitted = signal(false);
  // A required field is "missing" only once the user has tried to submit. Kept in
  // lockstep with the submit() guard below so what we show always matches what
  // actually blocks the create.
  readonly nameMissing = computed(() => this.submitted() && !this.name());
  readonly venueMissing = computed(() => this.submitted() && !this.venue_name());
  readonly startMissing = computed(() => this.submitted() && !this.start_date());

  submit() {
    // Surface validation before the guard: pressing Create with a field empty
    // must show why, not silently do nothing (the mobile "dead button" bug).
    this.submitted.set(true);
    if (!this.name() || !this.venue_name() || !this.start_date()) return;

    this.submitting.run(this.api.createEvent({
      name: this.name(),
      venue_name: this.venue_name(),
      start_date: this.start_date(),
      end_date: this.end_date(),
      genres: this.genres_input.split(',').map(g => g.trim()).filter(Boolean),
    }), {
      next: () => {
        this.store.refresh(); // invalidate the cached list so the new event shows
        this.router.navigate(['/admin/events']);
      },
    });
  }
}
