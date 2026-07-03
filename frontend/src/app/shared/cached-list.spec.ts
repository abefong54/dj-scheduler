import { Subject, of, throwError } from 'rxjs';
import { CachedList } from './cached-list';

describe('CachedList', () => {
  it('starts empty: value null, not loading, not errored', () => {
    const cache = new CachedList<number>(() => of([1, 2, 3]));
    expect(cache.value()).toBeNull();
    expect(cache.loading()).toBe(false);
    expect(cache.error()).toBe(false);
  });

  it('shows loading while the first fetch is in flight (nothing cached yet)', () => {
    const gate = new Subject<number[]>();
    const cache = new CachedList<number>(() => gate);
    cache.ensure();
    expect(cache.loading()).toBe(true);
    expect(cache.value()).toBeNull();
  });

  it('populates value and clears loading when the first fetch resolves', () => {
    const gate = new Subject<number[]>();
    const cache = new CachedList<number>(() => gate);
    cache.ensure();
    gate.next([1, 2, 3]);
    expect(cache.value()).toEqual([1, 2, 3]);
    expect(cache.loading()).toBe(false);
  });

  it('serves the cached value with NO loading spinner while revalidating', () => {
    const gate = new Subject<number[]>();
    const cache = new CachedList<number>(() => gate);
    cache.ensure();
    gate.next([1]);            // first load done

    const revalidate = new Subject<number[]>();
    (cache as unknown as { fetcher: () => unknown }).fetcher = () => revalidate;
    cache.ensure();            // second visit
    // cache present → no spinner even though a request is in flight
    expect(cache.loading()).toBe(false);
    expect(cache.value()).toEqual([1]);
    revalidate.next([1, 2]);   // fresh data arrives in the background
    expect(cache.value()).toEqual([1, 2]);
  });

  it('invalidate() drops the cache so the next ensure() spins again', () => {
    const cache = new CachedList<number>(() => of([1]));
    cache.ensure();
    expect(cache.value()).toEqual([1]);
    cache.invalidate();
    expect(cache.value()).toBeNull();
  });

  it('refresh() re-fetches and updates the cached value', () => {
    let batch = [1];
    const cache = new CachedList<number>(() => of(batch));
    cache.ensure();
    expect(cache.value()).toEqual([1]);
    batch = [1, 2, 3];
    cache.refresh();
    expect(cache.value()).toEqual([1, 2, 3]);
  });

  it('surfaces error on a failed FIRST load (nothing cached)', () => {
    const cache = new CachedList<number>(() => throwError(() => new Error('boom')));
    cache.ensure();
    expect(cache.error()).toBe(true);
    expect(cache.loading()).toBe(false);
    expect(cache.value()).toBeNull();
  });

  it('keeps stale data (and stays un-errored) when a background revalidation fails', () => {
    const gate = new Subject<number[]>();
    const cache = new CachedList<number>(() => gate);
    cache.ensure();
    gate.next([1, 2]);         // good first load

    (cache as unknown as { fetcher: () => unknown }).fetcher =
      () => throwError(() => new Error('network blip'));
    cache.refresh();           // background revalidation fails
    expect(cache.value()).toEqual([1, 2]);   // stale data preserved
    expect(cache.error()).toBe(false);       // no error UI over cached data
  });
});
