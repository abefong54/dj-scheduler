import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ApiService, DJ } from '../../services/api.service';
import { Challenge, ChallengesService } from '../../services/challenges.service';

/**
 * Teacher challenge authoring (Track H). Curate the school's core (badge) set and
 * author custom challenges assigned to a specific student. Backend CRUD lives at
 * /api/challenges; this is the console over it. Archiving is soft — earned badges
 * survive because they're snapshotted onto each passed submission.
 */
@Component({
  selector: 'app-admin-challenges',
  standalone: true,
  imports: [RouterLink],
  template: `
    <section class="mx-auto max-w-3xl px-6 py-8" style="color: var(--ink);">
      <header class="mb-6 flex items-center justify-between">
        <h1 class="font-display text-3xl font-bold tracking-tight">Challenges</h1>
        <a
          routerLink="/admin/review"
          class="btn btn-secondary px-4 py-2 text-sm"
          data-testid="review-link"
          >Review queue →</a
        >
      </header>

      <!-- Author form -->
      <div
        class="mb-8 rounded-[var(--radius-card)] border p-4"
        style="border-color: var(--border); background: var(--card);"
        data-testid="author-form"
      >
        <div class="mb-3 flex gap-2">
          <button
            class="btn px-4 py-2 text-sm"
            [class.btn-primary]="type() === 'core'"
            [class.btn-secondary]="type() !== 'core'"
            data-testid="type-core"
            (click)="type.set('core')"
          >
            Core (badge)
          </button>
          <button
            class="btn px-4 py-2 text-sm"
            [class.btn-primary]="type() === 'custom'"
            [class.btn-secondary]="type() !== 'custom'"
            data-testid="type-custom"
            (click)="type.set('custom')"
          >
            Custom (assigned)
          </button>
        </div>

        <input
          class="form-input mb-2"
          data-testid="challenge-title"
          placeholder="Title"
          [value]="title()"
          (input)="title.set(value($event))"
        />
        <input
          class="form-input mb-2"
          data-testid="challenge-desc"
          placeholder="Description"
          [value]="description()"
          (input)="description.set(value($event))"
        />

        @if (type() === 'core') {
          <div class="mb-2 flex gap-2">
            <input
              class="form-input"
              data-testid="badge-label"
              placeholder="Badge label"
              [value]="badgeLabel()"
              (input)="badgeLabel.set(value($event))"
            />
            <input
              class="form-input w-24"
              data-testid="badge-icon"
              placeholder="Icon (emoji)"
              [value]="badgeIcon()"
              (input)="badgeIcon.set(value($event))"
            />
          </div>
        } @else {
          <select
            class="form-input mb-2"
            data-testid="assign-student"
            [value]="assignedDjId()"
            (change)="assignedDjId.set(value($event))"
          >
            <option value="">Assign to a student…</option>
            @for (s of students(); track s.id) {
              <option [value]="s.id">{{ s.name }}</option>
            }
          </select>
        }

        @if (formError()) {
          <p class="mb-2 text-sm" data-testid="form-error" style="color:#dc2626;">
            {{ formError() }}
          </p>
        }
        <button
          class="btn btn-primary px-4 py-2 text-sm"
          data-testid="challenge-create"
          [disabled]="saving() || !canSubmit()"
          (click)="create()"
        >
          {{ saving() ? 'Saving…' : 'Add challenge' }}
        </button>
      </div>

      <!-- Existing challenges -->
      @if (loading()) {
        <p style="color: var(--muted);">Loading…</p>
      } @else {
        <div class="space-y-2" data-testid="challenge-list">
          @for (c of challenges(); track c.id) {
            <article
              class="flex items-center justify-between gap-3 rounded-[var(--radius-card)] border p-3"
              style="border-color: var(--border); background: var(--card);"
            >
              <div>
                <span
                  class="font-mono text-[10px] uppercase tracking-wide"
                  style="color: var(--muted);"
                  >{{ c.type }}</span
                >
                <h3 class="font-semibold">
                  {{ c.type === 'core' ? c.badge_icon + ' ' : '' }}{{ c.title }}
                </h3>
              </div>
              <button
                class="btn btn-secondary px-3 py-1 text-sm"
                data-testid="archive"
                (click)="archive(c.id)"
              >
                Archive
              </button>
            </article>
          } @empty {
            <p style="color: var(--muted);">No challenges yet — add your first above.</p>
          }
        </div>
      }
    </section>
  `,
})
export class AdminChallengesComponent {
  private svc = inject(ChallengesService);
  private api = inject(ApiService);

  readonly challenges = signal<Challenge[]>([]);
  readonly students = signal<DJ[]>([]);
  readonly loading = signal(true);

  readonly type = signal<'core' | 'custom'>('core');
  readonly title = signal('');
  readonly description = signal('');
  readonly badgeLabel = signal('');
  readonly badgeIcon = signal('');
  readonly assignedDjId = signal('');
  readonly saving = signal(false);
  readonly formError = signal<string | null>(null);

  readonly canSubmit = computed(() => {
    if (!this.title().trim()) return false;
    return this.type() === 'core' ? !!this.badgeLabel().trim() : !!this.assignedDjId();
  });

  constructor() {
    this.load();
  }

  private load(): void {
    this.loading.set(true);
    this.svc.listChallenges().subscribe({
      next: (c) => {
        this.challenges.set(c);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
    this.api
      .getDJs()
      .subscribe({ next: (d) => this.students.set(d.filter((x) => x.is_student !== false)) });
  }

  value(e: Event): string {
    return (e.target as HTMLInputElement | HTMLSelectElement).value;
  }

  create(): void {
    if (!this.canSubmit()) return;
    this.saving.set(true);
    this.formError.set(null);
    const done = {
      next: () => {
        this.saving.set(false);
        this.resetForm();
        this.load();
      },
      error: (e: { error?: { error?: string } }) => {
        this.saving.set(false);
        this.formError.set(e?.error?.error ?? 'Could not save that challenge.');
      },
    };
    if (this.type() === 'core') {
      this.svc
        .curateCore({
          title: this.title().trim(),
          description: this.description().trim(),
          badge_label: this.badgeLabel().trim(),
          badge_icon: this.badgeIcon().trim(),
        })
        .subscribe(done);
    } else {
      this.svc
        .authorChallenge({
          type: 'custom',
          title: this.title().trim(),
          description: this.description().trim(),
          assigned_dj_id: this.assignedDjId(),
        })
        .subscribe(done);
    }
  }

  archive(id: string): void {
    this.svc.archiveChallenge(id).subscribe({ next: () => this.load() });
  }

  private resetForm(): void {
    this.title.set('');
    this.description.set('');
    this.badgeLabel.set('');
    this.badgeIcon.set('');
    this.assignedDjId.set('');
  }
}
