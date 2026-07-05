import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { EMPTY, Observable, of } from 'rxjs';
import { environment } from '../../environments/environment';

// Student Challenges & Badges contracts (2026-07-05 sprint). These interfaces
// mirror the backend JSON (model.Challenge / model.Submission and the student
// session). SCAFFOLD: types are the contract between the backend (Tracks B/C) and
// frontend (Tracks F/G/H) tracks — do not change mid-wave without a heads-up.

/** A teacher-authored task. `core` = school-wide + badge; `custom` = assigned. */
export interface Challenge {
  id: string;
  school_id: string;
  type: 'core' | 'custom';
  title: string;
  description: string;
  skill: string;
  assigned_dj_id: string; // custom only, else ''
  badge_label: string;    // core only
  badge_icon: string;     // core only
  sort_order: number;
  created_by: string;
  archived: boolean;
  created_at: string;
}

/** A student's attempt at a challenge — feed item and progress driver. */
export interface Submission {
  id: string;
  school_id: string;
  challenge_id: string;
  dj_id: string;
  media_kind: string;      // 'embed' (future: 'file')
  media_provider: string;  // youtube | soundcloud | mixcloud
  media_ref: string;
  caption: string;
  status: 'posted' | 'passed' | 'needs_work';
  review_note: string;
  earned_badge_label: string; // snapshot at pass, core only
  earned_badge_icon: string;
  reviewed_by: string;
  reviewed_at: string;
  created_at: string;
}

/** The logged-in student's session, resolved from the student JWT. */
export interface StudentSession {
  dj_id: string;
  school_id: string;
  name: string;
}

/** Input for posting/re-working an attempt (student). */
export interface PostAttemptInput {
  challenge_id: string;
  media_url: string;
  caption?: string;
}

/** Input for a teacher review decision. */
export interface ReviewInput {
  status: 'passed' | 'needs_work';
  review_note?: string;
}

/**
 * ChallengesService is the typed client for the challenge/submission endpoints.
 * SCAFFOLD: every method is a stub returning an empty placeholder — Track F fills
 * the real HTTP calls against `this.base`. Signatures are the contract.
 */
@Injectable({ providedIn: 'root' })
export class ChallengesService {
  private http = inject(HttpClient);
  private base = environment.apiUrl;

  // --- student surfaces ---

  /** List the logged-in student's challenges (assigned custom + picked core). */
  listMyChallenges(): Observable<Challenge[]> {
    return of([]);
  }

  /** Browse the school's core (badge-bearing) challenges. */
  browseCore(): Observable<Challenge[]> {
    return of([]);
  }

  /** Post (or re-work) an attempt against a challenge. */
  postAttempt(input: PostAttemptInput): Observable<Submission> {
    return EMPTY;
  }

  /** The school feed of passed submissions, newest first. */
  listFeed(): Observable<Submission[]> {
    return of([]);
  }

  // --- teacher surfaces ---

  /** Author a custom challenge for one student. */
  authorChallenge(input: Partial<Challenge>): Observable<Challenge> {
    return EMPTY;
  }

  /** Add or edit a core (badge) challenge in the school's set. */
  curateCore(input: Partial<Challenge>): Observable<Challenge> {
    return EMPTY;
  }

  /** Pass or return a submission with an optional note. */
  reviewSubmission(id: string, input: ReviewInput): Observable<Submission> {
    return EMPTY;
  }
}
