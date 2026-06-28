import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-event-new',
  standalone: true,
  imports: [FormsModule, RouterLink, TranslatePipe],
  templateUrl: './event-new.component.html',
  styleUrl: './event-new.component.css',
})
export class EventNewComponent {
  private api = inject(ApiService);
  private router = inject(Router);
  private translate = inject(TranslateService);

  name = signal('');
  venue_name = signal('');
  start_date = signal('');
  end_date = signal('');
  genres_input = '';

  submit() {
    if (!this.name() || !this.venue_name() || !this.start_date()) return;

    this.api.createEvent({
      name: this.name(),
      venue_name: this.venue_name(),
      start_date: this.start_date(),
      end_date: this.end_date(),
      genres: this.genres_input.split(',').map(g => g.trim()).filter(Boolean),
    }).subscribe(() => {
      this.router.navigate(['/admin/events']);
    });
  }
}
