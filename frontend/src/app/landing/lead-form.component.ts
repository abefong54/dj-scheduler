import { Component, computed, inject, signal } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';
import { LeadService } from './lead.service';

type LeadFormState = 'idle' | 'submitting' | 'success' | 'error';

@Component({
  selector: 'app-lead-form',
  standalone: true,
  imports: [TranslatePipe],
  templateUrl: './lead-form.component.html',
  styleUrl: './lead-form.component.css',
})
export class LeadFormComponent {
  private readonly leads = inject(LeadService);

  readonly name = signal('');
  readonly organization = signal('');
  readonly email = signal('');
  readonly message = signal('');
  readonly state = signal<LeadFormState>('idle');

  private readonly emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  readonly valid = computed(
    () => this.name().trim().length > 0 && this.emailRe.test(this.email().trim()),
  );

  /** Read an input/textarea value from a DOM event (signal-bound inputs). */
  val(e: Event): string {
    return (e.target as HTMLInputElement | HTMLTextAreaElement).value;
  }

  submit(): void {
    if (!this.valid() || this.state() === 'submitting') return;
    this.state.set('submitting');
    this.leads
      .submit({
        name: this.name().trim(),
        organization: this.organization().trim(),
        email: this.email().trim(),
        message: this.message().trim(),
      })
      .subscribe({
        next: () => this.state.set('success'),
        error: () => this.state.set('error'),
      });
  }
}
