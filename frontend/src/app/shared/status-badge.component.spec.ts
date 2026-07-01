import { Component } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';
import { StatusBadgeComponent, BadgeStatus } from './status-badge.component';

@Component({
  standalone: true,
  imports: [StatusBadgeComponent],
  template: `<app-status-badge [status]="status" [label]="label" [testId]="testId" />`,
})
class HostComponent {
  status: BadgeStatus = 'confirmed';
  label: string | null = null;
  testId: string | null = null;
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
    it(`renders the ${s.status} state with its tint class, glyph and label`, () => {
      fixture.componentInstance.status = s.status;
      fixture.detectChanges();
      const el = badge();
      expect(el).toBeTruthy();
      expect(el.classList).toContain(s.cls);
      // EL-078: status must carry an inline glyph AND a text label — never
      // colour alone — so it survives red/green colour-blindness.
      const glyph = el.querySelector('.status-badge-glyph');
      expect(glyph).toBeTruthy();
      expect(glyph!.querySelector('path, rect, circle')).toBeTruthy();
      const label = el.querySelector('.status-badge-label');
      expect(label).toBeTruthy();
      // No translations loaded → ngx-translate echoes the key, proving wiring.
      expect(label!.textContent).toContain(s.key);
    });
  }

  it('renders a distinct glyph shape per state (icon differs, not just colour)', () => {
    // A signature of which primitive shapes each glyph contains — different per
    // state, proving the icon (not only the colour) distinguishes the status.
    const signature = (status: BadgeStatus) => {
      const f = TestBed.createComponent(HostComponent);
      f.componentInstance.status = status;
      f.detectChanges();
      const glyph = f.nativeElement.querySelector('.status-badge-glyph')!;
      return {
        path: glyph.querySelectorAll('path').length,
        rect: glyph.querySelectorAll('rect').length,
        circle: glyph.querySelectorAll('circle').length,
      };
    };
    expect(signature('confirmed')).toEqual({ path: 1, rect: 0, circle: 0 });
    expect(signature('pending')).toEqual({ path: 0, rect: 2, circle: 0 });
    expect(signature('declined')).toEqual({ path: 1, rect: 1, circle: 0 });
    expect(signature('live')).toEqual({ path: 0, rect: 0, circle: 1 });
  });

  // EL-081: an optional label override lets the pill carry a custom label (e.g. a
  // certification name) while keeping the icon + label discipline (never colour
  // alone). Reused by the DJ roster to render "Cleared" certification chips.
  it('renders the label override instead of the translated status copy, keeping the glyph', () => {
    fixture.componentInstance.status = 'confirmed';
    fixture.componentInstance.label = 'House';
    fixture.detectChanges();
    const el = badge();
    expect(el.querySelector('.status-badge-label')!.textContent).toContain('House');
    expect(el.textContent).not.toContain('badge.confirmed');
    expect(el.querySelector('svg.status-badge-glyph')).toBeTruthy();
  });

  it('applies the optional testId to the pill root', () => {
    fixture.componentInstance.testId = 'dj-cert-42';
    fixture.detectChanges();
    expect(badge().getAttribute('data-testid')).toBe('dj-cert-42');
  });
});
