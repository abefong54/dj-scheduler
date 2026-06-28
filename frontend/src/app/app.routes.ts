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
    path: 'admin/events/:id/slots/new',
    canActivate: [authGuard],
    loadComponent: () => import('./admin/events/slot-new.component')
      .then(m => m.SlotNewComponent),
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
    path: 'events/:id',
    loadComponent: () => import('./schedule/schedule.component')
      .then(m => m.ScheduleComponent),
  },
  { path: '', redirectTo: 'admin/events', pathMatch: 'full' },
];
