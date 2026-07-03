import { Component } from '@angular/core';
import { TestBed } from '@angular/core/testing';

import { TooltipDirective } from './tooltip.directive';

@Component({
  standalone: true,
  imports: [TooltipDirective],
  template: `<button [appTooltip]="text">x</button>`,
})
class HostComponent {
  text = 'Edit DJ';
}

// EL-090: the tooltip renders on <body> (so it escapes clipping ancestors like the
// scrolling data table) on hover, and is removed on leave.
describe('TooltipDirective (EL-090)', () => {
  afterEach(() => {
    document.body.querySelectorAll('.app-tooltip').forEach(el => el.remove());
  });

  function host(text?: string) {
    const fixture = TestBed.createComponent(HostComponent);
    if (text !== undefined) fixture.componentInstance.text = text;
    fixture.detectChanges();
    return fixture.nativeElement.querySelector('button') as HTMLButtonElement;
  }

  it('shows the tooltip text on mouseenter and removes it on mouseleave', () => {
    const btn = host('Edit DJ');

    btn.dispatchEvent(new MouseEvent('mouseenter'));
    expect(document.body.querySelector('.app-tooltip')?.textContent).toBe('Edit DJ');

    btn.dispatchEvent(new MouseEvent('mouseleave'));
    expect(document.body.querySelector('.app-tooltip')).toBeNull();
  });

  it('shows nothing when the tooltip text is empty', () => {
    const btn = host('');
    btn.dispatchEvent(new MouseEvent('mouseenter'));
    expect(document.body.querySelector('.app-tooltip')).toBeNull();
  });
});
