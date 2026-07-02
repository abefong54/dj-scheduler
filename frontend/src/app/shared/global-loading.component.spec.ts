import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { GlobalLoadingComponent } from './global-loading.component';
import { LoadingService } from '../services/loading.service';

describe('GlobalLoadingComponent (EL-087)', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    TestBed.configureTestingModule({ imports: [GlobalLoadingComponent] });
  });
  afterEach(() => vi.useRealTimers());

  const indicator = (f: { nativeElement: HTMLElement }) =>
    f.nativeElement.querySelector('[data-testid="global-loading"]');

  it('appears only after the 150ms grace period while loading', async () => {
    const loading = TestBed.inject(LoadingService);
    const f = TestBed.createComponent(GlobalLoadingComponent);
    f.detectChanges();
    expect(indicator(f)).toBeNull();

    loading.start();
    await vi.advanceTimersByTimeAsync(149);
    f.detectChanges();
    expect(indicator(f)).toBeNull();

    await vi.advanceTimersByTimeAsync(1);
    f.detectChanges();
    expect(indicator(f)).not.toBeNull();
  });

  it('does not flash for a request that finishes within the grace period', async () => {
    const loading = TestBed.inject(LoadingService);
    const f = TestBed.createComponent(GlobalLoadingComponent);
    f.detectChanges();

    loading.start();
    await vi.advanceTimersByTimeAsync(50);
    loading.stop();
    await vi.advanceTimersByTimeAsync(200);
    f.detectChanges();
    expect(indicator(f)).toBeNull();
  });
});
