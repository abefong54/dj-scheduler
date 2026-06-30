import { Component } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';
import { StatusBadgeComponent, BadgeStatus } from './status-badge.component';

@Component({
  standalone: true,
  imports: [StatusBadgeComponent],
  template: `<app-status-badge [status]="status" />`,
})
class HostComponent {
  status: BadgeStatus = 'confirmed';
}

describe('StatusBadgeComponent', () => {
  let fixture: ComponentFixture<HostComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HostComponent],
      providers: [provideTranslateService()],
    }).compileComponents();
    fixture = TestBed.createComponent(HostComponent);
  });

  const badge = () =>
    fixture.nativeElement.querySelector('.status-badge') as HTMLElement;

  const STATES: { status: BadgeStatus; cls: string; key: string }[] = [
    { status: 'confirmed', cls: 'status-confirmed', key: 'badge.confirmed' },
    { status: 'pending', cls: 'status-pending', key: 'badge.pending' },
    { status: 'declined', cls: 'status-declined', key: 'badge.declined' },
    { status: 'live', cls: 'status-live', key: 'badge.live' },
  ];

  for (const s of STATES) {
    it(`renders the ${s.status} state with its tint class, dot and label`, () => {
      fixture.componentInstance.status = s.status;
      fixture.detectChanges();
      const el = badge();
      expect(el).toBeTruthy();
      expect(el.classList).toContain(s.cls);
      expect(el.querySelector('.status-badge-dot')).toBeTruthy();
      // No translations loaded → ngx-translate echoes the key, proving wiring.
      expect(el.textContent).toContain(s.key);
    });
  }
});
