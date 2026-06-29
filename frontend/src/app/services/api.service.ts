import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

export interface DJ {
  id: string;
  name: string;
  genre_tags: string[];
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
}

// Shape of GET /api/events/:id/public — the unauthenticated schedule payload.
export interface PublicSchedule {
  event: Event;
  stages: Stage[];
  slots: Slot[];
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

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // DJs
  getDJs() { return this.http.get<DJ[]>(`${this.base}/api/djs`); }
  createDJ(d: Pick<DJ, 'name' | 'genre_tags'>) { return this.http.post<DJ>(`${this.base}/api/djs`, d); }
  deleteDJ(id: string) { return this.http.delete(`${this.base}/api/djs/${id}`); }

  // Events
  getEvents() { return this.http.get<Event[]>(`${this.base}/api/events`); }
  getEvent(id: string) { return this.http.get<Event>(`${this.base}/api/events/${id}`); }
  // Unauthenticated schedule for the public/shareable view (no Bearer token needed).
  getPublicSchedule(id: string) { return this.http.get<PublicSchedule>(`${this.base}/api/events/${id}/public`); }
  createEvent(e: Omit<Event, 'id'>) { return this.http.post<Event>(`${this.base}/api/events`, e); }
  deleteEvent(id: string) { return this.http.delete(`${this.base}/api/events/${id}`); }
  cloneEvent(id: string) { return this.http.post<Event>(`${this.base}/api/events/${id}/clone`, {}); }

  // DJ portal (token-gated, no auth header)
  getDJPortal(token: string) {
    return this.http.get<DJPortalResponse>(`${this.base}/api/dj/portal`, { params: { token } });
  }
  confirmSlot(slotId: string, confirmation: 'confirmed' | 'flagged', token: string) {
    return this.http.patch<{ id: string; dj_confirmation: DJConfirmation }>(
      `${this.base}/api/dj/portal/slots/${slotId}`, { confirmation }, { params: { token } });
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
    return this.http.put<Slot>(`${this.base}/api/events/${eventId}/slots/${slotId}`, s);
  }
  deleteSlot(eventId: string, slotId: string) {
    return this.http.delete(`${this.base}/api/events/${eventId}/slots/${slotId}`);
  }
}
