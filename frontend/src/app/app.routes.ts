import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

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
  {
    path: 'admin/events',
    canActivate: [authGuard],
    loadComponent: () => import('./admin/events/events-list.component')
      .then(m => m.EventsListComponent),
  },
  {
    path: 'admin/events/new',
    canActivate: [authGuard],
    loadComponent: () => import('./admin/events/event-new.component')
      .then(m => m.EventNewComponent),
  },
  {
    path: 'admin/events/:id',
    canActivate: [authGuard],
    loadComponent: () => import('./admin/events/event-detail.component')
      .then(m => m.EventDetailComponent),
  },
  {
    path: 'admin/djs',
    canActivate: [authGuard],
    loadComponent: () => import('./admin/djs/djs.component')
      .then(m => m.DjsComponent),
  },
  {
    path: 'admin/performance',
    canActivate: [authGuard],
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
