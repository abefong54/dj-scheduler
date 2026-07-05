import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { AdminReviewComponent } from './review.component';
import { Challenge, ChallengesService, Submission } from '../../services/challenges.service';
import { ApiService, DJ } from '../../services/api.service';

function ch(over: Partial<Challenge>): Challenge {
  return {
    id: 'c',
    school_id: 's',
    type: 'core',
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
function sub(over: Partial<Submission>): Submission {
  return {
    id: 'sub',
    school_id: 's',
    challenge_id: 'c1',
    dj_id: 'dj1',
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
const DJ_ADA: DJ = { id: 'dj1', name: 'Ada', genre_tags: [] };

function build(pending: Submission[] = [sub({ id: 'sub1', caption: 'my mix' })]) {
  const svc = {
    listPending: vi.fn(() => of(pending)),
    listChallenges: vi.fn(() => of([ch({ id: 'c1', title: 'Beatmatcher' })])),
    reviewSubmission: vi.fn(() => of(sub({}))),
  };
  const api = { getDJs: vi.fn(() => of([DJ_ADA])) };
  TestBed.configureTestingModule({
    imports: [AdminReviewComponent],
    providers: [
      provideRouter([]),
      { provide: ChallengesService, useValue: svc },
      { provide: ApiService, useValue: api },
    ],
  });
  const fixture = TestBed.createComponent(AdminReviewComponent);
  fixture.detectChanges();
  return { fixture, svc, el: fixture.nativeElement as HTMLElement };
}

describe('AdminReviewComponent (Track H)', () => {
  it('renders pending items with the student name, challenge title, and embed', () => {
    const { el } = build();
    const item = el.querySelector('[data-testid="review-item"]');
    expect(item?.querySelector('[data-testid="review-student"]')?.textContent).toContain('Ada');
    expect(item?.textContent).toContain('Beatmatcher');
    expect(item?.querySelector('[data-testid="media-embed-frame"]')?.tagName.toLowerCase()).toBe(
      'iframe',
    );
  });

  it('passes a submission', () => {
    const { svc, el } = build();
    (el.querySelector('[data-testid="pass"]') as HTMLButtonElement).click();
    expect(svc.reviewSubmission).toHaveBeenCalledWith('sub1', { status: 'passed' });
  });

  it('sends a submission back with a needs-work note', () => {
    const { fixture, svc, el } = build();
    (el.querySelector('[data-testid="open-needs-work"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    const note = el.querySelector('[data-testid="needs-work-note"]') as HTMLInputElement;
    note.value = 'tighten the transition';
    note.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    (el.querySelector('[data-testid="needs-work-submit"]') as HTMLButtonElement).click();
    expect(svc.reviewSubmission).toHaveBeenCalledWith('sub1', {
      status: 'needs_work',
      review_note: 'tighten the transition',
    });
  });

  it('shows an empty state when the queue is clear', () => {
    const { el } = build([]);
    expect(el.querySelector('[data-testid="review-empty"]')).toBeTruthy();
  });
});
