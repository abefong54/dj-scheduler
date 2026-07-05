import { Component, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ApiService, DJ } from '../../services/api.service';
import { Challenge, ChallengesService, Submission } from '../../services/challenges.service';
import { MediaEmbedComponent } from '../../shared/media-embed/media-embed.component';

/**
 * Teacher review queue (Track H): posted submissions awaiting sign-off. Each shows
 * the student's embedded mix, who submitted it, and against which challenge, with
 * Pass / Needs-work actions. Passing a core challenge snapshots the badge and
 * pushes the mix to the feed (both handled server-side). Challenge titles and
 * student names are joined client-side from the roster + challenge list.
 */
@Component({
  selector: 'app-admin-review',
  standalone: true,
  imports: [RouterLink, MediaEmbedComponent],
  template: `
    <section class="mx-auto max-w-2xl px-6 py-8" style="color: var(--ink);">
      <header class="mb-6 flex items-center justify-between">
        <h1 class="font-display text-3xl font-bold tracking-tight">Review queue</h1>
        <a
          routerLink="/admin/challenges"
          class="btn btn-secondary px-4 py-2 text-sm"
          data-testid="challenges-link"
          >← Challenges</a
        >
      </header>

      @if (loading()) {
        <p style="color: var(--muted);">Loading…</p>
      } @else {
        <div class="space-y-6" data-testid="review-list">
          @for (s of pending(); track s.id) {
            <article
              class="rounded-[var(--radius-card)] border p-4"
              style="border-color: var(--border); background: var(--card);"
              data-testid="review-item"
            >
              <div class="mb-2 text-sm">
                <span class="font-semibold" data-testid="review-student">{{
                  studentName(s.dj_id)
                }}</span>
                <span style="color: var(--muted);"> · {{ challengeTitle(s.challenge_id) }}</span>
              </div>
              <app-media-embed [media]="s" />
              @if (s.caption) {
                <p class="mt-2 text-sm">{{ s.caption }}</p>
              }

              <div class="mt-3 flex flex-wrap items-center gap-2">
                <button
                  class="btn btn-primary px-4 py-2 text-sm"
                  data-testid="pass"
                  [disabled]="busy() === s.id"
                  (click)="pass(s)"
                >
                  Pass
                </button>
                <button
                  class="btn btn-secondary px-4 py-2 text-sm"
                  data-testid="open-needs-work"
                  (click)="openNote(s.id)"
                >
                  Needs work
                </button>
              </div>

              @if (noteFor() === s.id) {
                <div class="mt-2 flex gap-2" data-testid="needs-work-form">
                  <input
                    class="form-input"
                    data-testid="needs-work-note"
                    placeholder="What to fix (optional)"
                    [value]="note()"
                    (input)="note.set(value($event))"
                  />
                  <button
                    class="btn btn-secondary px-4 py-2 text-sm"
                    data-testid="needs-work-submit"
                    [disabled]="busy() === s.id"
                    (click)="needsWork(s)"
                  >
                    Send back
                  </button>
                </div>
              }
            </article>
          } @empty {
            <p data-testid="review-empty" style="color: var(--muted);">
              Nothing to review — the queue is clear.
            </p>
          }
        </div>
      }
    </section>
  `,
})
export class AdminReviewComponent {
  private svc = inject(ChallengesService);
  private api = inject(ApiService);

  readonly pending = signal<Submission[]>([]);
  readonly loading = signal(true);
  readonly busy = signal<string | null>(null);
  readonly noteFor = signal<string | null>(null);
  readonly note = signal('');

  private readonly challenges = signal<Challenge[]>([]);
  private readonly students = signal<DJ[]>([]);

  constructor() {
    this.load();
  }

  private load(): void {
    this.loading.set(true);
    this.svc.listPending().subscribe({
      next: (p) => {
        this.pending.set(p);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
    this.svc.listChallenges().subscribe({ next: (c) => this.challenges.set(c) });
    this.api.getDJs().subscribe({ next: (d) => this.students.set(d) });
  }

  value(e: Event): string {
    return (e.target as HTMLInputElement).value;
  }

  challengeTitle(id: string): string {
    return this.challenges().find((c) => c.id === id)?.title ?? 'Challenge';
  }

  studentName(djId: string): string {
    return this.students().find((d) => d.id === djId)?.name ?? 'Student';
  }

  openNote(id: string): void {
    this.noteFor.set(id);
    this.note.set('');
  }

  pass(s: Submission): void {
    this.review(s.id, { status: 'passed' });
  }

  needsWork(s: Submission): void {
    this.review(s.id, { status: 'needs_work', review_note: this.note().trim() || undefined });
  }

  private review(
    id: string,
    input: { status: 'passed' | 'needs_work'; review_note?: string },
  ): void {
    this.busy.set(id);
    this.svc.reviewSubmission(id, input).subscribe({
      next: () => {
        this.busy.set(null);
        this.noteFor.set(null);
        this.load();
      },
      error: () => this.busy.set(null),
    });
  }
}
