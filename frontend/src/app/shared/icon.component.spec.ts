import { ComponentFixture, TestBed } from '@angular/core/testing';
import { IconComponent, ICON_NAMES, IconName } from './icon.component';

describe('IconComponent', () => {
  let fixture: ComponentFixture<IconComponent>;

  function render(name: IconName) {
    fixture = TestBed.createComponent(IconComponent);
    fixture.componentRef.setInput('name', name);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [IconComponent],
    }).compileComponents();
  });

  it('renders an inline svg', () => {
    const el = render('trash');
    expect(el.querySelector('svg')).not.toBeNull();
  });

  it('marks the svg aria-hidden so the label carries meaning', () => {
    const el = render('trash');
    expect(el.querySelector('svg')?.getAttribute('aria-hidden')).toBe('true');
  });

  it('exposes the selected icon via data-icon', () => {
    const el = render('edit');
    expect(el.querySelector('svg')?.getAttribute('data-icon')).toBe('edit');
  });

  it('renders a non-empty glyph (at least one path) for every supported name', () => {
    for (const name of ICON_NAMES) {
      const el = render(name);
      const svg = el.querySelector('svg');
      expect(svg, `svg missing for "${name}"`).not.toBeNull();
      expect(
        svg!.querySelectorAll('path, polyline, line, circle, rect').length,
        `no glyph shape rendered for "${name}"`,
      ).toBeGreaterThan(0);
    }
  });
});
