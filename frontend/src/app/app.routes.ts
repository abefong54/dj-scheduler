import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';
import { hasSchoolGuard, onboardingGuard } from './guards/onboarding.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./auth/login.component')
      .then(m => m.LoginComponent),
  },
  {
    path: 'auth/callback',
    loadComponent: () => import('./auth/auth-callback.component')
      .then(m => m.AuthCallbackComponent),
  },
  // EL-106: onboarding — authenticated users who have no school yet land here.
  // hasSchoolGuard bounces users who already have a school back to /admin/events.
  {
    path: 'onboarding',
    canActivate: [authGuard, hasSchoolGuard],
    loadComponent: () => import('./onboarding/onboarding.component')
      .then(m => m.OnboardingComponent),
  },
  {
    path: 'admin/events',
    canActivate: [authGuard, onboardingGuard],
    loadComponent: () => import('./admin/events/events-list.component')
      .then(m => m.EventsListComponent),
  },
  {
    path: 'admin/events/new',
    canActivate: [authGuard, onboardingGuard],
    loadComponent: () => import('./admin/events/event-new.component')
      .then(m => m.EventNewComponent),
  },
  {
    path: 'admin/events/:id',
    canActivate: [authGuard, onboardingGuard],
    loadComponent: () => import('./admin/events/event-detail.component')
      .then(m => m.EventDetailComponent),
  },
  {
    path: 'admin/djs',
    canActivate: [authGuard, onboardingGuard],
    loadComponent: () => import('./admin/djs/djs.component')
      .then(m => m.DjsComponent),
  },
  {
    path: 'admin/performance',
    canActivate: [authGuard, onboardingGuard],
    loadComponent: () => import('./admin/performance/performance.component')
      .then(m => m.PerformanceComponent),
  },
  {
    path: 'events/:id',
    loadComponent: () => import('./schedule/schedule.component')
      .then(m => m.ScheduleComponent),
  },
  {
    // Public, token-gated DJ portal (no auth guard) — the link DJs receive.
    path: 'dj/portal',
    loadComponent: () => import('./dj-portal/dj-portal.component')
      .then(m => m.DJPortalComponent),
  },
  {
    // Public per-DJ share card (no auth guard) — the link a DJ shares (EL-049).
    path: 'card/:slotId',
    loadComponent: () => import('./card/card.component')
      .then(m => m.CardComponent),
  },
  // Student features (2026-07-05 sprint) — SCAFFOLD placeholder routes. Auth
  // gating on the student JWT is a Track E concern; shells load unguarded for now.
  {
    path: 'student/login',
    loadComponent: () => import('./student/student-login.component')
      .then(m => m.StudentLoginComponent),
  },
  {
    path: 'student/challenges',
    loadComponent: () => import('./student/student-challenges.component')
      .then(m => m.StudentChallengesComponent),
  },
  {
    path: 'student/feed',
    loadComponent: () => import('./student/student-feed.component')
      .then(m => m.StudentFeedComponent),
  },
  {
    // Public Soundcheck marketing landing — the branded front door prospects
    // see before signing in (EL-075). Authenticated users can navigate into the
    // console from the top bar; auth callbacks still land on /admin/events.
    path: '',
    pathMatch: 'full',
    loadComponent: () => import('./landing/landing.component')
      .then(m => m.LandingComponent),
  },
];
