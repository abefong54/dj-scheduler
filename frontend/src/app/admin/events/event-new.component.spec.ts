import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { vi } from 'vitest';
import { provideTranslateService } from '@ngx-translate/core';

import { EventNewComponent } from './event-new.component';
import { ApiService } from '../../services/api.service';

describe('EventNewComponent (EL-065 Console restyle)', () => {
  let fixture: ComponentFixture<EventNewComponent>;
  let component: EventNewComponent;
  const root = () => fixture.nativeElement as HTMLElement;

  const createEvent = vi.fn(() => of({}));
  const apiMock = { createEvent } as unknown as ApiService;

  beforeEach(async () => {
    createEvent.mockClear();
    await TestBed.configureTestingModule({
      imports: [EventNewComponent],
      providers: [
        provideRouter([{ path: '**', children: [] }]),
        provideTranslateService(),
        { provide: ApiService, useValue: apiMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(EventNewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('renders inside the AdminShell with the Console form card', () => {
    expect(root().querySelector('app-admin-shell')).toBeTruthy();
    expect(root().querySelector('.shell-sidebar')).toBeTruthy();
    expect(root().querySelector('form.console-card')).toBeTruthy();
  });

  it('marks the Events nav item active in the shell', () => {
    const active = root().querySelector('.shell-nav-active');
    expect(active?.getAttribute('data-nav')).toBe('events');
  });

  it('keeps every create-event field bound to the existing model', () => {
    for (const id of [
      'event-name-input',
      'event-venue-input',
      'event-start-date',
      'event-end-date',
      'event-genres-input',
    ]) {
      expect(root().querySelector(`[data-testid="${id}"]`)).toBeTruthy();
    }
  });

  it('renders a violet primary "Create event" and a secondary Cancel', () => {
    expect(
      root().querySelector('[data-testid="event-create-btn"].console-btn-primary'),
    ).toBeTruthy();
    expect(
      root().querySelector('[data-testid="event-cancel-btn"].console-btn-secondary'),
    ).toBeTruthy();
  });

  it('does not submit when required fields are empty (validation unchanged)', () => {
    component.submit();
    expect(createEvent).not.toHaveBeenCalled();
  });

  it('submits the existing payload when required fields are filled', () => {
    component.name.set('Summer Fest');
    component.venue_name.set('The Warehouse');
    component.start_date.set('2026-07-01');
    component.end_date.set('2026-07-02');
    component.genres_input = 'House, Techno';

    component.submit();

    expect(createEvent).toHaveBeenCalledWith({
      name: 'Summer Fest',
      venue_name: 'The Warehouse',
      start_date: '2026-07-01',
      end_date: '2026-07-02',
      genres: ['House', 'Techno'],
    });
  });
});
