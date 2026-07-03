import { TestBed } from '@angular/core/testing';
import { BreakpointService } from './breakpoint.service';

// Minimal matchMedia stub whose `matches` we control and whose registered
// `change` listener we can fire to simulate the viewport crossing 768px.
function stubMatchMedia(initial: boolean) {
  let listener: ((e: MediaQueryListEvent) => void) | null = null;
  const mql = {
    matches: initial,
    addEventListener: (_: string, cb: (e: MediaQueryListEvent) => void) => {
      listener = cb;
    },
    removeEventListener: () => {
      listener = null;
    },
  };
  window.matchMedia = (() => mql) as unknown as typeof window.matchMedia;
  return {
    emit: (matches: boolean) => listener?.({ matches } as MediaQueryListEvent),
    hasListener: () => listener !== null,
  };
}

describe('BreakpointService (EL-094)', () => {
  const original = window.matchMedia;
  afterEach(() => {
    window.matchMedia = original;
  });

  it('seeds compact() from the initial media match', () => {
    stubMatchMedia(true);
    const svc = TestBed.inject(BreakpointService);
    expect(svc.compact()).toBe(true);
  });

  it('reacts to viewport changes', () => {
    const mm = stubMatchMedia(false);
    const svc = TestBed.inject(BreakpointService);
    expect(svc.compact()).toBe(false);
    mm.emit(true);
    expect(svc.compact()).toBe(true);
    mm.emit(false);
    expect(svc.compact()).toBe(false);
  });
});
