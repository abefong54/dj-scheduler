import { Component, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ChallengesService, Submission } from '../services/challenges.service';
import { MediaEmbedComponent } from '../shared/media-embed/media-embed.component';

/**
 * School feed (Track G): the stream of teacher-passed submissions, newest first —
 * the venue-free "payoff moment". Each item renders the student's mix inline via
 * the media-embed seam, with its caption and (for core passes) the earned badge.
 */
@Component({
  selector: 'app-student-feed',
  standalone: true,
  imports: [RouterLink, MediaEmbedComponent],
  template: `
    <section class="mx-auto max-w-2xl px-6 py-8" style="color: var(--ink);">
      <header class="mb-8 flex items-center justify-between">
        <h1 class="font-display text-3xl font-bold tracking-tight">Feed</h1>
        <a
          routerLink="/student/challenges"
          class="btn btn-secondary px-4 py-2 text-sm"
          data-testid="challenges-link"
          >← My Challenges</a
        >
      </header>

      @if (loading()) {
        <p style="color: var(--muted);">Loading…</p>
      } @else if (error()) {
        <p data-testid="load-error" style="color: var(--muted);">
          Couldn't load the feed. Try again shortly.
        </p>
      } @else {
        <div class="space-y-6" data-testid="feed">
          @for (s of feed(); track s.id) {
            <article
              class="rounded-[var(--radius-card)] border p-4"
              style="border-color: var(--border); background: var(--card);"
              data-testid="feed-item"
            >
              <app-media-embed [media]="s" />
              @if (s.caption) {
                <p class="mt-3 text-sm">{{ s.caption }}</p>
              }
              @if (s.earned_badge_label) {
                <span
                  class="mt-3 inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs"
                  style="border-color: var(--border);"
                  data-testid="feed-badge"
                >
                  <span aria-hidden="true">{{ s.earned_badge_icon }}</span
                  >{{ s.earned_badge_label }}
                </span>
              }
            </article>
          } @empty {
            <p data-testid="feed-empty" style="color: var(--muted);">
              No passed sets yet — be the first to land one.
            </p>
          }
        </div>
      }
    </section>
  `,
})
export class StudentFeedComponent {
  private svc = inject(ChallengesService);

  readonly feed = signal<Submission[]>([]);
  readonly loading = signal(true);
  readonly error = signal(false);

  constructor() {
    this.svc.listFeed().subscribe({
      next: (f) => {
        this.feed.set(f);
        this.loading.set(false);
      },
      error: () => {
        this.error.set(true);
        this.loading.set(false);
      },
    });
  }
}
