import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { StudentFeedComponent } from './student-feed.component';
import { ChallengesService, Submission } from '../services/challenges.service';

function submission(over: Partial<Submission>): Submission {
  return {
    id: 'sub',
    school_id: 's',
    challenge_id: 'c',
    dj_id: 'dj',
    media_kind: 'embed',
    media_provider: 'youtube',
    media_ref: 'abc',
    caption: '',
    status: 'passed',
    review_note: '',
    earned_badge_label: '',
    earned_badge_icon: '',
    reviewed_by: '',
    reviewed_at: '',
    created_at: '',
    ...over,
  };
}

function build(feed: Submission[]) {
  const svc = { listFeed: vi.fn(() => of(feed)) };
  TestBed.configureTestingModule({
    imports: [StudentFeedComponent],
    providers: [provideRouter([]), { provide: ChallengesService, useValue: svc }],
  });
  const fixture = TestBed.createComponent(StudentFeedComponent);
  fixture.detectChanges();
  return fixture.nativeElement as HTMLElement;
}

describe('StudentFeedComponent (Track G)', () => {
  it('renders passed submissions with an embedded player, caption, and badge', () => {
    const el = build([
      submission({
        id: 'f1',
        media_provider: 'youtube',
        media_ref: 'dQw4w9WgXcQ',
        caption: 'my first set',
        earned_badge_label: 'Beatmatcher',
        earned_badge_icon: '🥁',
      }),
    ]);
    const item = el.querySelector('[data-testid="feed-item"]');
    expect(item).toBeTruthy();
    expect(item?.querySelector('[data-testid="media-embed-frame"]')?.tagName.toLowerCase()).toBe(
      'iframe',
    );
    expect(item?.textContent).toContain('my first set');
    expect(el.querySelector('[data-testid="feed-badge"]')?.textContent).toContain('Beatmatcher');
  });

  it('shows an empty state when nothing has passed yet', () => {
    const el = build([]);
    expect(el.querySelector('[data-testid="feed-empty"]')).toBeTruthy();
  });

  it('shows a load error when the feed fetch fails', () => {
    const svc = { listFeed: vi.fn(() => ({ subscribe: (h: { error: () => void }) => h.error() })) };
    TestBed.configureTestingModule({
      imports: [StudentFeedComponent],
      providers: [provideRouter([]), { provide: ChallengesService, useValue: svc }],
    });
    const fixture = TestBed.createComponent(StudentFeedComponent);
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-testid="load-error"]'),
    ).toBeTruthy();
  });
});
