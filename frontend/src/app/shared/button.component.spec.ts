import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ButtonComponent } from './button.component';

describe('ButtonComponent', () => {
  let fixture: ComponentFixture<ButtonComponent>;

  function render(inputs: Record<string, unknown> = {}) {
    fixture = TestBed.createComponent(ButtonComponent);
    for (const [k, v] of Object.entries(inputs)) {
      fixture.componentRef.setInput(k, v);
    }
    fixture.detectChanges();
    const host = fixture.nativeElement as HTMLElement;
    return { host, button: host.querySelector('button')! };
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ButtonComponent],
    }).compileComponents();
  });

  it('renders a native button element', () => {
    const { button } = render();
    expect(button).not.toBeNull();
  });

  it('applies base + default variant/size classes', () => {
    const { button } = render();
    expect(button.className).toContain('btn');
    expect(button.className).toContain('btn-secondary');
    expect(button.className).toContain('btn-md');
  });

  it('maps variant and size to the matching classes', () => {
    const { button } = render({ variant: 'danger', size: 'icon' });
    expect(button.className).toContain('btn-danger');
    expect(button.className).toContain('btn-icon');
  });

  it('passes through the button type', () => {
    const { button } = render({ type: 'submit' });
    expect(button.getAttribute('type')).toBe('submit');
  });

  it('reflects the disabled input', () => {
    const { button } = render({ disabled: true });
    expect(button.disabled).toBe(true);
  });

  it('forwards testId to data-testid', () => {
    const { button } = render({ testId: 'dj-delete-1' });
    expect(button.getAttribute('data-testid')).toBe('dj-delete-1');
  });

  it('renders the icon when an icon name is provided', () => {
    const { button } = render({ icon: 'trash' });
    expect(button.querySelector('svg[data-icon="trash"]')).not.toBeNull();
  });

  it('shows the label text in normal (non-icon-only) mode', () => {
    const { button } = render({ label: 'Save' });
    expect(button.textContent?.trim()).toBe('Save');
    expect(button.getAttribute('aria-label')).toBeNull();
  });

  it('hides the text and exposes the label as aria-label + title when iconOnly', () => {
    const { button } = render({ icon: 'trash', iconOnly: true, label: 'Delete' });
    expect(button.textContent?.trim()).toBe('');
    expect(button.getAttribute('aria-label')).toBe('Delete');
    expect(button.getAttribute('title')).toBe('Delete');
  });

  it('lets ariaLabel override the label for the accessible name', () => {
    const { button } = render({ icon: 'link', iconOnly: true, label: 'Copy', ariaLabel: 'Copy portal link' });
    expect(button.getAttribute('aria-label')).toBe('Copy portal link');
  });

  it('stretches to full width when fullWidth is set', () => {
    const { button } = render({ fullWidth: true, label: 'Share' });
    expect(button.className).toContain('w-full');
  });

  it('is not full width by default', () => {
    const { button } = render({ label: 'Share' });
    expect(button.className).not.toContain('w-full');
  });

  it('bubbles clicks to the host so a parent (click) handler fires', () => {
    const { host, button } = render();
    let clicks = 0;
    host.addEventListener('click', () => clicks++);
    button.click();
    expect(clicks).toBe(1);
  });

  it('emits no click when disabled', () => {
    const { host, button } = render({ disabled: true });
    let clicks = 0;
    host.addEventListener('click', () => clicks++);
    button.click();
    expect(clicks).toBe(0);
  });
});
