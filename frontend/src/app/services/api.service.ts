import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

export interface DJ {
  id: string;
  name: string;
  genre_tags: string[];
  // EL-019: genres the DJ is certified to perform, and student vs graduate.
  // Optional so existing DJ literals/tests stay valid; the API always returns them.
  certifications?: string[];
  is_student?: boolean;
}

export interface Event {
  id: string;
  name: string;
  venue_name: string;
  start_date: string; // 'YYYY-MM-DD'
  end_date: string;   // 'YYYY-MM-DD'
  genres: string[];
}

export interface Stage {
  id: string;
  event_id: string;
  name: string;
  color: string;
  display_order: number;
}

export interface Slot {
  id: string;
  event_id: string;
  stage_id: string;
  stage_name: string;
  dj_id: string;
  dj_name: string;
  genre: string;
  slot_date: string;   // 'YYYY-MM-DD'
  start_time: string;  // 'HH:MM'
  end_time: string;    // 'HH:MM'
  notes: string;
  dj_confirmation: DJConfirmation; // DJ portal response, null = no response yet
}

// Shape of GET /api/events/:id/public — the unauthenticated schedule payload.
export interface PublicSchedule {
  event: Event;
  stages: Stage[];
  slots: Slot[];
}

// Shape of GET /api/slots/:id/public — a single booking plus its event, for the
// public per-DJ share card (EL-049). No auth; exposes only already-public slot data.
export interface PublicSlot {
  slot: Slot;
  event: Event;
}

// A DJ's portal response is "confirmed", "flagged", or null (no response yet).
export type DJConfirmation = 'confirmed' | 'flagged' | null;

export interface DJPortalSlot {
  id: string;
  event_id: string;
  event_name: string;
  stage_name: string;
  genre: string;
  slot_date: string;
  start_time: string;
  end_time: string;
  notes: string;
  dj_confirmation: DJConfirmation;
}

export interface DJPortalResponse {
  dj: { id: string; name: string; genre_tags: string[] };
  slots: DJPortalSlot[];
}

// Performance aggregation (EL-043/044): "reps" = slots played. total_minutes is
// real stage time (sets crossing midnight count their full length). last_played
// is 'YYYY-MM-DD' or '' when the DJ has never played.
export interface GenreStat {
  genre: string;
  reps: number;
  total_minutes: number;
}

export interface DJPerformance {
  dj_id: string;
  dj_name: string;
  reps: number;
  total_minutes: number;
  last_played: string;
  by_genre: GenreStat[];
}

export interface RosterPerformance {
  dj_id: string;
  dj_name: string;
  is_student: boolean;
  reps: number;
  total_minutes: number;
  last_played: string;
}

// Optional window/event filters shared by the roster + under-served endpoints.
export interface PerformanceFilter {
  eventId?: string;
  from?: string; // 'YYYY-MM-DD'
  to?: string;
}

// portalAuth builds the Authorization header carrying a DJ portal token (EL-038).
function portalAuth(token: string) {
  return { Authorization: `Bearer ${token}` };
}

// performanceParams maps the optional window/event filter to query params, omitting
// empties so the backend's "unset" branches apply.
function performanceParams(filter: PerformanceFilter): Record<string, string> {
  const params: Record<string, string> = {};
  if (filter.eventId) params['event_id'] = filter.eventId;
  if (filter.from) params['from'] = filter.from;
  if (filter.to) params['to'] = filter.to;
  return params;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // DJs
  getDJs() { return this.http.get<DJ[]>(`${this.base}/api/djs`); }
  // EL-019: optional ?certified_for= / ?ready=true filters on the roster.
  getDJsFiltered(opts: { certifiedFor?: string; ready?: boolean }) {
    const params: Record<string, string> = {};
    if (opts.certifiedFor) params['certified_for'] = opts.certifiedFor;
    if (opts.ready) params['ready'] = 'true';
    return this.http.get<DJ[]>(`${this.base}/api/djs`, { params });
  }
  createDJ(d: Pick<DJ, 'name' | 'genre_tags'>) { return this.http.post<DJ>(`${this.base}/api/djs`, d); }
  // EL-020: update a DJ's name, genres, certifications, and student status (PATCH).
  updateDJ(id: string, d: Pick<DJ, 'name' | 'genre_tags' | 'certifications' | 'is_student'>) {
    return this.http.patch<DJ>(`${this.base}/api/djs/${id}`, d);
  }
  deleteDJ(id: string) { return this.http.delete(`${this.base}/api/djs/${id}`); }
  // Mint (or regenerate) a DJ's personal portal link.
  generateDJPortalToken(djId: string) {
    return this.http.post<{ portal_url: string; expires_at: string }>(`${this.base}/api/djs/${djId}/token`, {});
  }

  // Performance (EL-043/044). Routes registered by the backend performance handler:
  // GET /api/djs/:id/performance, GET /api/performance, GET /api/performance/underserved.
  getDJPerformance(id: string) {
    return this.http.get<DJPerformance>(`${this.base}/api/djs/${id}/performance`);
  }
  getPerformance(filter: PerformanceFilter = {}) {
    return this.http.get<RosterPerformance[]>(`${this.base}/api/performance`, { params: performanceParams(filter) });
  }
  getUnderserved(threshold?: number, filter: PerformanceFilter = {}) {
    const params = performanceParams(filter);
    if (threshold != null) params['threshold'] = String(threshold);
    return this.http.get<RosterPerformance[]>(`${this.base}/api/performance/underserved`, { params });
  }

  // Events
  getEvents() { return this.http.get<Event[]>(`${this.base}/api/events`); }
  getEvent(id: string) { return this.http.get<Event>(`${this.base}/api/events/${id}`); }
  // Unauthenticated schedule for the public/shareable view (no Bearer token needed).
  getPublicSchedule(id: string) { return this.http.get<PublicSchedule>(`${this.base}/api/events/${id}/public`); }
  // Unauthenticated single slot + its event, for the per-DJ share card (EL-049).
  getPublicSlot(slotId: string) { return this.http.get<PublicSlot>(`${this.base}/api/slots/${slotId}/public`); }
  createEvent(e: Omit<Event, 'id'>) { return this.http.post<Event>(`${this.base}/api/events`, e); }
  deleteEvent(id: string) { return this.http.delete(`${this.base}/api/events/${id}`); }
  cloneEvent(id: string) { return this.http.post<Event>(`${this.base}/api/events/${id}/clone`, {}); }

  // DJ portal (token-gated). EL-038: the portal token travels in the
  // Authorization header, never a query param, so it can't leak via browser
  // history, Referer headers, or access logs.
  getDJPortal(token: string) {
    return this.http.get<DJPortalResponse>(`${this.base}/api/dj/portal`, { headers: portalAuth(token) });
  }
  confirmSlot(slotId: string, confirmation: 'confirmed' | 'flagged', token: string) {
    return this.http.patch<{ id: string; dj_confirmation: DJConfirmation }>(
      `${this.base}/api/dj/portal/slots/${slotId}`, { confirmation }, { headers: portalAuth(token) });
  }

  // Stages
  getStages(eventId: string) { return this.http.get<Stage[]>(`${this.base}/api/events/${eventId}/stages`); }
  createStage(eventId: string, s: Pick<Stage, 'name' | 'color'>) {
    return this.http.post<Stage>(`${this.base}/api/events/${eventId}/stages`, s);
  }
  deleteStage(eventId: string, stageId: string) {
    return this.http.delete(`${this.base}/api/events/${eventId}/stages/${stageId}`);
  }

  // Slots
  getSlots(eventId: string) { return this.http.get<Slot[]>(`${this.base}/api/events/${eventId}/slots`); }
  createSlot(eventId: string, s: Pick<Slot, 'stage_id' | 'dj_id' | 'genre' | 'slot_date' | 'start_time' | 'end_time' | 'notes'>) {
    return this.http.post<Slot>(`${this.base}/api/events/${eventId}/slots`, s);
  }
  updateSlot(eventId: string, slotId: string, s: Pick<Slot, 'stage_id' | 'dj_id' | 'genre' | 'slot_date' | 'start_time' | 'end_time' | 'notes'>) {
    return this.http.patch<Slot>(`${this.base}/api/events/${eventId}/slots/${slotId}`, s);
  }
  deleteSlot(eventId: string, slotId: string) {
    return this.http.delete(`${this.base}/api/events/${eventId}/slots/${slotId}`);
  }
}
