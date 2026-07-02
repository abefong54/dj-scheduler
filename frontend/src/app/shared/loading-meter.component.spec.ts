import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoadingMeterComponent } from './loading-meter.component';

describe('LoadingMeterComponent', () => {
  let fixture: ComponentFixture<LoadingMeterComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoadingMeterComponent],
    }).compileComponents();
    fixture = TestBed.createComponent(LoadingMeterComponent);
  });

  const root = () => fixture.nativeElement as HTMLElement;

  it('renders VU-meter bars ("levels, not fades")', () => {
    fixture.detectChanges();
    expect(root().querySelector('.vu-loader')).toBeTruthy();
    expect(root().querySelectorAll('.vu-loader-bar').length).toBe(4);
  });

  it('exposes a status role for assistive tech', () => {
    fixture.detectChanges();
    expect(root().querySelector('[role="status"]')).toBeTruthy();
  });

  it('renders an optional label when provided', () => {
    fixture.componentRef.setInput('label', 'Loading…');
    fixture.detectChanges();
    const label = root().querySelector('.vu-loader-label');
    expect(label?.textContent).toContain('Loading…');
    expect(root().querySelector('.vu-loader')?.getAttribute('aria-label')).toBe(
      'Loading…',
    );
  });

  it('omits the label element when no label is set', () => {
    fixture.detectChanges();
    expect(root().querySelector('.vu-loader-label')).toBeNull();
  });
});
