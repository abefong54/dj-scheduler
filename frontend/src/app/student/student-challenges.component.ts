import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Challenge, ChallengesService, Submission } from '../services/challenges.service';

/**
 * Student portal home (Track G). Three surfaces over the same data: earned
 * badges (passed core challenges, read from the snapshot on each submission),
 * "My Challenges" (assigned custom + picked-up core, each with an inline
 * post-attempt form), and "Browse Core" to pick up a new badge challenge. Posting
 * an attempt is the payoff loop's entry point — paste a SoundCloud/Mixcloud/
 * YouTube link and it goes to the teacher for sign-off.
 */
@Component({
  selector: 'app-student-challenges',
  standalone: true,
  imports: [RouterLink],
  template: `
    <section class="mx-auto max-w-3xl px-6 py-8" style="color: var(--ink);">
      <header class="mb-8 flex items-center justify-between">
        <h1 class="font-display text-3xl font-bold tracking-tight">My Challenges</h1>
        <a
          routerLink="/student/feed"
          class="btn btn-secondary px-4 py-2 text-sm"
          data-testid="feed-link"
          >Feed →</a
        >
      </header>

      @if (loading()) {
        <p style="color: var(--muted);">Loading…</p>
      } @else if (error()) {
        <p data-testid="load-error" style="color: var(--muted);">
          Couldn't load your challenges. Try again shortly.
        </p>
      } @else {
        @if (earnedBadges().length) {
          <div class="mb-8" data-testid="badges">
            <h2
              class="mb-3 font-mono text-xs uppercase tracking-[0.16em]"
              style="color: var(--muted);"
            >
              Badges earned
            </h2>
            <div class="flex flex-wrap gap-2">
              @for (b of earnedBadges(); track b.id) {
                <span
                  class="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-sm"
                  style="border-color: var(--border); background: var(--card);"
                  data-testid="badge"
                >
                  <span aria-hidden="true">{{ b.earned_badge_icon }}</span
                  >{{ b.earned_badge_label }}
                </span>
              }
            </div>
          </div>
        }

        <div class="space-y-3" data-testid="my-challenges">
          @for (c of myChallenges(); track c.id) {
            <article
              class="rounded-[var(--radius-card)] border p-4"
              style="border-color: var(--border); background: var(--card);"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="font-semibold">{{ c.title }}</h3>
                  <p class="mt-1 text-sm" style="color: var(--muted);">{{ c.description }}</p>
                </div>
                @if (submissionFor(c.id); as sub) {
                  <span
                    class="shrink-0 font-mono text-[11px] uppercase tracking-wide"
                    data-testid="status"
                    [style.color]="statusColor(sub.status)"
                  >
                    {{ statusLabel(sub.status) }}
                  </span>
                }
              </div>

              @if (submissionFor(c.id)?.review_note; as note) {
                <p class="mt-2 text-sm" style="color: var(--muted);">Teacher note: {{ note }}</p>
              }

              @if (activeChallengeId() === c.id) {
                <div class="mt-3 space-y-2" data-testid="post-form">
                  <input
                    class="form-input"
                    data-testid="post-url"
                    placeholder="Paste your SoundCloud / Mixcloud / YouTube link"
                    [value]="draftUrl()"
                    (input)="draftUrl.set(value($event))"
                  />
                  <input
                    class="form-input"
                    data-testid="post-caption"
                    placeholder="Caption (optional)"
                    [value]="draftCaption()"
                    (input)="draftCaption.set(value($event))"
                  />
                  @if (postError()) {
                    <p class="text-sm" data-testid="post-error" style="color:#dc2626;">
                      {{ postError() }}
                    </p>
                  }
                  <div class="flex gap-2">
                    <button
                      class="btn btn-primary px-4 py-2 text-sm"
                      data-testid="post-submit"
                      [disabled]="posting() || !draftUrl().trim()"
                      (click)="submitPost()"
                    >
                      {{ posting() ? 'Posting…' : 'Post attempt' }}
                    </button>
                    <button class="btn btn-secondary px-4 py-2 text-sm" (click)="cancelPost()">
                      Cancel
                    </button>
                  </div>
                </div>
              } @else {
                <button
                  class="btn btn-secondary mt-3 px-4 py-2 text-sm"
                  data-testid="open-post"
                  (click)="openPost(c.id)"
                >
                  {{ submissionFor(c.id) ? 'Re-post attempt' : 'Post attempt' }}
                </button>
              }
            </article>
          } @empty {
            <p style="color: var(--muted);">
              No challenges yet — pick one from Browse Core below or ask your teacher.
            </p>
          }
        </div>

        @if (browseable().length) {
          <div class="mt-10" data-testid="browse-core">
            <h2
              class="mb-3 font-mono text-xs uppercase tracking-[0.16em]"
              style="color: var(--muted);"
            >
              Browse core — earn a badge
            </h2>
            <div class="space-y-3">
              @for (c of browseable(); track c.id) {
                <article
                  class="flex items-center justify-between gap-3 rounded-[var(--radius-card)] border p-4"
                  style="border-color: var(--border); background: var(--card);"
                >
                  <div>
                    <h3 class="font-semibold">{{ c.badge_icon }} {{ c.title }}</h3>
                    <p class="mt-1 text-sm" style="color: var(--muted);">{{ c.description }}</p>
                  </div>
                  <button
                    class="btn btn-secondary shrink-0 px-4 py-2 text-sm"
                    data-testid="pick-core"
                    (click)="openPost(c.id)"
                  >
                    Try this
                  </button>
                </article>
              }
            </div>
          </div>
        }
      }
    </section>
  `,
})
export class StudentChallengesComponent {
  private svc = inject(ChallengesService);

  readonly myChallenges = signal<Challenge[]>([]);
  readonly coreChallenges = signal<Challenge[]>([]);
  readonly submissions = signal<Submission[]>([]);
  readonly loading = signal(true);
  readonly error = signal(false);

  readonly activeChallengeId = signal<string | null>(null);
  readonly draftUrl = signal('');
  readonly draftCaption = signal('');
  readonly posting = signal(false);
  readonly postError = signal<string | null>(null);

  /** Earned badges = passed submissions of core challenges (snapshot on the row). */
  readonly earnedBadges = computed(() =>
    this.submissions().filter((s) => s.status === 'passed' && s.earned_badge_label),
  );

  /** Core challenges not already assigned/picked, so Browse Core shows fresh ones. */
  readonly browseable = computed(() => {
    const mine = new Set(this.myChallenges().map((c) => c.id));
    return this.coreChallenges().filter((c) => !mine.has(c.id));
  });

  constructor() {
    this.load();
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(false);
    this.svc.listMyChallenges().subscribe({
      next: (c) => {
        this.myChallenges.set(c);
        this.loading.set(false);
      },
      error: () => {
        this.error.set(true);
        this.loading.set(false);
      },
    });
    this.svc.browseCore().subscribe({ next: (c) => this.coreChallenges.set(c) });
    this.svc.listMySubmissions().subscribe({ next: (s) => this.submissions.set(s) });
  }

  submissionFor(challengeId: string): Submission | undefined {
    return this.submissions().find((s) => s.challenge_id === challengeId);
  }

  value(e: Event): string {
    return (e.target as HTMLInputElement).value;
  }

  statusLabel(status: Submission['status']): string {
    return status === 'needs_work' ? 'needs work' : status;
  }

  statusColor(status: Submission['status']): string {
    if (status === 'passed') return '#16a34a';
    if (status === 'needs_work') return '#d97706';
    return 'var(--muted)';
  }

  openPost(challengeId: string): void {
    this.activeChallengeId.set(challengeId);
    this.draftUrl.set('');
    this.draftCaption.set('');
    this.postError.set(null);
  }

  cancelPost(): void {
    this.activeChallengeId.set(null);
  }

  submitPost(): void {
    const id = this.activeChallengeId();
    const url = this.draftUrl().trim();
    if (!id || !url) return;
    this.posting.set(true);
    this.postError.set(null);
    this.svc
      .postAttempt({
        challenge_id: id,
        media_url: url,
        caption: this.draftCaption().trim() || undefined,
      })
      .subscribe({
        next: () => {
          this.posting.set(false);
          this.activeChallengeId.set(null);
          this.load();
        },
        error: (e: { error?: { error?: string } }) => {
          this.posting.set(false);
          this.postError.set(
            e?.error?.error ??
              'Could not post that link — use a SoundCloud, Mixcloud, or YouTube URL.',
          );
        },
      });
  }
}
