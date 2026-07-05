import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { of } from 'rxjs';

import { AdminChallengesComponent } from './challenges.component';
import { Challenge, ChallengesService } from '../../services/challenges.service';
import { ApiService, DJ } from '../../services/api.service';

function ch(over: Partial<Challenge>): Challenge {
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
const DJ_ADA: DJ = { id: 'dj1', name: 'Ada', genre_tags: [], is_student: true };

function build(over: Record<string, unknown> = {}) {
  const svc = {
    listChallenges: vi.fn(() =>
      of([ch({ id: 'c1', type: 'core', title: 'Beatmatcher', badge_icon: '🥁' })]),
    ),
    curateCore: vi.fn(() => of(ch({}))),
    authorChallenge: vi.fn(() => of(ch({}))),
    archiveChallenge: vi.fn(() => of(undefined)),
    ...over,
  };
  const api = { getDJs: vi.fn(() => of([DJ_ADA])) };
  TestBed.configureTestingModule({
    imports: [AdminChallengesComponent],
    providers: [
      provideRouter([]),
      provideTranslateService(),
      { provide: ChallengesService, useValue: svc },
      { provide: ApiService, useValue: api },
    ],
  });
  const fixture = TestBed.createComponent(AdminChallengesComponent);
  fixture.detectChanges();
  return { fixture, svc, el: fixture.nativeElement as HTMLElement };
}

function type(el: HTMLElement, testid: string, val: string) {
  const input = el.querySelector(`[data-testid="${testid}"]`) as
    HTMLInputElement | HTMLSelectElement;
  input.value = val;
  input.dispatchEvent(new Event(input.tagName === 'SELECT' ? 'change' : 'input'));
}

describe('AdminChallengesComponent (Track H)', () => {
  it('renders inside the admin shell so the sidebar is present', () => {
    const { el } = build();
    // The page must be wrapped in <app-admin-shell> — otherwise it renders bare
    // with no sidebar (the nav the "Challenges" button links to).
    expect(el.querySelector('app-admin-shell')).toBeTruthy();
    expect(el.querySelector('.shell-sidebar')).toBeTruthy();
    // The active nav item should be Challenges.
    expect(
      el.querySelector('.shell-nav-item.shell-nav-active .shell-nav-label')?.textContent,
    ).toContain('shell.nav.challenges');
    // The review-queue link is projected into the topbar actions slot.
    expect(el.querySelector('.shell-actions-group [data-testid="review-link"]')).toBeTruthy();
  });

  it('lists existing challenges', () => {
    const { el } = build();
    expect(el.querySelector('[data-testid="challenge-list"]')?.textContent).toContain(
      'Beatmatcher',
    );
  });

  it('creates a core challenge with a badge', () => {
    const { fixture, svc, el } = build();
    type(el, 'challenge-title', 'Clean Cut');
    type(el, 'badge-label', 'Clean Cut');
    fixture.detectChanges();
    (el.querySelector('[data-testid="challenge-create"]') as HTMLButtonElement).click();
    expect(svc.curateCore).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'Clean Cut', badge_label: 'Clean Cut' }),
    );
  });

  it('authors a custom challenge assigned to a student', () => {
    const { fixture, svc, el } = build();
    (el.querySelector('[data-testid="type-custom"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    type(el, 'challenge-title', 'EQ transitions');
    type(el, 'assign-student', 'dj1');
    fixture.detectChanges();
    (el.querySelector('[data-testid="challenge-create"]') as HTMLButtonElement).click();
    expect(svc.authorChallenge).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'custom', title: 'EQ transitions', assigned_dj_id: 'dj1' }),
    );
  });

  it('archives a challenge', () => {
    const { svc, el } = build();
    (el.querySelector('[data-testid="archive"]') as HTMLButtonElement).click();
    expect(svc.archiveChallenge).toHaveBeenCalledWith('c1');
  });
});
