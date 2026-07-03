import { Directive, ElementRef, HostListener, OnDestroy, Renderer2, inject, input } from '@angular/core';

/**
 * Lightweight hover/focus tooltip for icon-only controls.
 *
 * It renders a fixed-position element on `document.body` so it escapes clipping
 * ancestors — a pure-CSS `::after` tooltip is cut off inside the horizontally
 * scrolling data table (`.dt-table-scroll { overflow-x: auto }`), which clips
 * vertical overflow. The accessible name still comes from the host's
 * `aria-label`; this is the visible affordance only. (EL-090)
 */
@Directive({
  selector: '[appTooltip]',
  standalone: true,
})
export class TooltipDirective implements OnDestroy {
  /** Tooltip text. Empty string → no tooltip is shown. */
  readonly appTooltip = input<string>('');

  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly r = inject(Renderer2);
  private tip: HTMLElement | null = null;

  @HostListener('mouseenter')
  @HostListener('focus')
  show() {
    const text = this.appTooltip();
    if (!text || this.tip) return;

    const el = this.r.createElement('div') as HTMLElement;
    this.r.addClass(el, 'app-tooltip');
    this.r.setAttribute(el, 'role', 'tooltip');
    this.r.setProperty(el, 'textContent', text);
    this.r.appendChild(document.body, el);

    const rect = this.host.nativeElement.getBoundingClientRect();
    this.r.setStyle(el, 'left', `${rect.left + rect.width / 2}px`);
    this.r.setStyle(el, 'top', `${rect.top}px`);
    this.tip = el;
  }

  @HostListener('mouseleave')
  @HostListener('blur')
  hide() {
    if (this.tip) {
      this.r.removeChild(document.body, this.tip);
      this.tip = null;
    }
  }

  ngOnDestroy() {
    this.hide();
  }
}
