import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import {
  ChallengesService,
  Challenge,
  Submission,
  PostAttemptInput,
  ReviewInput,
} from './challenges.service';

const SUBMISSION: Submission = {
  id: 'sub-1',
  school_id: 'sch-1',
  challenge_id: 'ch-1',
  dj_id: 'dj-1',
  media_kind: 'embed',
  media_provider: 'youtube',
  media_ref: 'dQw4w9WgXcQ',
  caption: 'my mix',
  status: 'posted',
  review_note: '',
  earned_badge_label: '',
  earned_badge_icon: '',
  reviewed_by: '',
  reviewed_at: '',
  created_at: '2026-07-05T00:00:00Z',
};

const CHALLENGE: Challenge = {
  id: 'ch-1',
  school_id: 'sch-1',
  type: 'core',
  title: 'Beatmatch',
  description: '',
  skill: 'mixing',
  assigned_dj_id: '',
  badge_label: 'Beatmatcher',
  badge_icon: '🎧',
  sort_order: 0,
  created_by: 'org-1',
  archived: false,
  created_at: '2026-07-05T00:00:00Z',
};

describe('ChallengesService', () => {
  let svc: ChallengesService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ChallengesService, provideHttpClient(), provideHttpClientTesting()],
    });
    svc = TestBed.inject(ChallengesService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  // --- student surfaces ---

  it('listMyChallenges GETs /api/student/challenges', () => {
    let result: Challenge[] | undefined;
    svc.listMyChallenges().subscribe((r) => (result = r));
    const req = httpMock.expectOne('/api/student/challenges');
    expect(req.request.method).toBe('GET');
    req.flush([CHALLENGE]);
    expect(result).toEqual([CHALLENGE]);
  });

  it('browseCore GETs /api/student/challenges/core', () => {
    svc.browseCore().subscribe();
    const req = httpMock.expectOne('/api/student/challenges/core');
    expect(req.request.method).toBe('GET');
    req.flush([]);
  });

  it('postAttempt POSTs the attempt body to /api/student/submissions', () => {
    const input: PostAttemptInput = {
      challenge_id: 'ch-1',
      media_url: 'https://youtu.be/dQw4w9WgXcQ',
      caption: 'my mix',
    };
    let result: Submission | undefined;
    svc.postAttempt(input).subscribe((r) => (result = r));
    const req = httpMock.expectOne('/api/student/submissions');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(input);
    req.flush(SUBMISSION);
    expect(result).toEqual(SUBMISSION);
  });

  it('listMySubmissions GETs /api/student/submissions', () => {
    svc.listMySubmissions().subscribe();
    const req = httpMock.expectOne('/api/student/submissions');
    expect(req.request.method).toBe('GET');
    req.flush([SUBMISSION]);
  });

  it('listFeed GETs /api/student/feed', () => {
    svc.listFeed().subscribe();
    const req = httpMock.expectOne('/api/student/feed');
    expect(req.request.method).toBe('GET');
    req.flush([]);
  });

  // --- teacher surfaces ---

  it('authorChallenge POSTs the challenge body to /api/challenges', () => {
    const input: Partial<Challenge> = { type: 'custom', title: 'Scratch', assigned_dj_id: 'dj-1' };
    svc.authorChallenge(input).subscribe();
    const req = httpMock.expectOne('/api/challenges');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(input);
    req.flush(CHALLENGE);
  });

  it('updateChallenge PUTs to /api/challenges/:id', () => {
    const input: Partial<Challenge> = { title: 'Renamed' };
    svc.updateChallenge('ch-1', input).subscribe();
    const req = httpMock.expectOne('/api/challenges/ch-1');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual(input);
    req.flush(CHALLENGE);
  });

  it('curateCore POSTs to /api/challenges with type=core forced', () => {
    svc.curateCore({ title: 'Beatmatch', badge_label: 'Beatmatcher' }).subscribe();
    const req = httpMock.expectOne('/api/challenges');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      type: 'core',
      title: 'Beatmatch',
      badge_label: 'Beatmatcher',
    });
    req.flush(CHALLENGE);
  });

  it('reviewSubmission POSTs to /api/submissions/:id/review', () => {
    const input: ReviewInput = { status: 'passed', review_note: 'nice' };
    let result: Submission | undefined;
    svc.reviewSubmission('sub-1', input).subscribe((r) => (result = r));
    const req = httpMock.expectOne('/api/submissions/sub-1/review');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(input);
    req.flush({ ...SUBMISSION, status: 'passed' });
    expect(result?.status).toBe('passed');
  });
});
