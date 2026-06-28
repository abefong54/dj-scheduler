import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'admin/events',
    loadComponent: () => import('./admin/events/events-list.component')
      .then(m => m.EventsListComponent),
  },
  {
    path: 'admin/events/new',
    loadComponent: () => import('./admin/events/event-new.component')
      .then(m => m.EventNewComponent),
  },
  {
    path: 'admin/events/:id/slots/new',
    loadComponent: () => import('./admin/events/slot-new.component')
      .then(m => m.SlotNewComponent),
  },
  {
    path: 'admin/events/:id',
    loadComponent: () => import('./admin/events/event-detail.component')
      .then(m => m.EventDetailComponent),
  },
  {
    path: 'admin/djs',
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
