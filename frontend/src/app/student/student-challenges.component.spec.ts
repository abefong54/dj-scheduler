import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { StudentChallengesComponent } from './student-challenges.component';
import { Challenge, ChallengesService, Submission } from '../services/challenges.service';

function challenge(over: Partial<Challenge>): Challenge {
  return {
    id: 'c',
    school_id: 's',
    type: 'custom',
    title: '',
    description: '',
    skill: '',
    assigned_dj_id: '',
    badge_label: '',
    badge_icon: '',
    sort_order: 0,
    created_by: '',
    archived: false,
    created_at: '',
    ...over,
  };
}
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
    status: 'posted',
    review_note: '',
    earned_badge_label: '',
    earned_badge_icon: '',
    reviewed_by: '',
    reviewed_at: '',
    created_at: '',
    ...over,
  };
}

const CUSTOM = challenge({
  id: 'ch1',
  type: 'custom',
  title: 'EQ transitions',
  description: 'work on EQ',
});
const CORE = challenge({
  id: 'ch2',
  type: 'core',
  title: 'Beatmatcher',
  badge_label: 'Beatmatcher',
  badge_icon: '🥁',
});
const PASSED = submission({
  id: 's1',
  challenge_id: 'ch2',
  status: 'passed',
  earned_badge_label: 'Beatmatcher',
  earned_badge_icon: '🥁',
});

function build(over: Partial<Record<keyof ChallengesService, unknown>> = {}) {
  const svc = {
    listMyChallenges: vi.fn(() => of([CUSTOM])),
    browseCore: vi.fn(() => of([CORE])),
    listMySubmissions: vi.fn(() => of([PASSED])),
    postAttempt: vi.fn(() => of(submission({}))),
    ...over,
  };
  TestBed.configureTestingModule({
    imports: [StudentChallengesComponent],
    providers: [provideRouter([]), { provide: ChallengesService, useValue: svc }],
  });
  const fixture = TestBed.createComponent(StudentChallengesComponent);
  fixture.detectChanges();
  return { fixture, svc, el: fixture.nativeElement as HTMLElement };
}

describe('StudentChallengesComponent (Track G)', () => {
  it('lists my challenges and earned badges', () => {
    const { el } = build();
    expect(el.querySelector('[data-testid="my-challenges"]')?.textContent).toContain(
      'EQ transitions',
    );
    expect(el.querySelector('[data-testid="badges"]')?.textContent).toContain('Beatmatcher');
  });

  it('offers un-picked core challenges under Browse Core', () => {
    const { el } = build();
    expect(el.querySelector('[data-testid="browse-core"]')?.textContent).toContain('Beatmatcher');
  });

  it('posts an attempt for a challenge and refreshes', () => {
    const { fixture, svc, el } = build();
    (el.querySelector('[data-testid="open-post"]') as HTMLButtonElement).click();
    fixture.detectChanges();

    const url = el.querySelector('[data-testid="post-url"]') as HTMLInputElement;
    url.value = 'https://youtu.be/abc123';
    url.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    (el.querySelector('[data-testid="post-submit"]') as HTMLButtonElement).click();
    expect(svc.postAttempt).toHaveBeenCalledWith({
      challenge_id: 'ch1',
      media_url: 'https://youtu.be/abc123',
      caption: undefined,
    });
    // reloads after a successful post (initial load + reload = 2 calls)
    expect(svc.listMyChallenges).toHaveBeenCalledTimes(2);
  });

  it('surfaces a backend rejection message on a bad link', () => {
    const { fixture, el } = build({
      postAttempt: vi.fn(() => ({
        subscribe: (h: { error: (e: unknown) => void }) =>
          h.error({ error: { error: 'unsupported provider' } }),
      })),
    });
    (el.querySelector('[data-testid="open-post"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    const url = el.querySelector('[data-testid="post-url"]') as HTMLInputElement;
    url.value = 'https://vimeo.com/1';
    url.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    (el.querySelector('[data-testid="post-submit"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(el.querySelector('[data-testid="post-error"]')?.textContent).toContain(
      'unsupported provider',
    );
  });

  it('shows a load error when the challenge fetch fails', () => {
    const { el } = build({
      listMyChallenges: vi.fn(() => ({ subscribe: (h: { error: () => void }) => h.error() })),
    });
    expect(el.querySelector('[data-testid="load-error"]')).toBeTruthy();
  });
});
