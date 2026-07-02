import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

export interface LeadPayload {
  name: string;
  organization: string;
  email: string;
  message: string;
}

@Injectable({ providedIn: 'root' })
export class LeadService {
  private readonly http = inject(HttpClient);

  /** POST a lead to the public capture endpoint. */
  submit(payload: LeadPayload): Observable<void> {
    return this.http.post('/api/leads', payload).pipe(map(() => void 0));
  }
}
