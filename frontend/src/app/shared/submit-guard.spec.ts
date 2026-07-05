import { Observable, Subject } from 'rxjs';
import { SubmitGuard } from './submit-guard';

describe('SubmitGuard (BUG-008)', () => {
  it('runs the first call and flips pending for its lifetime', () => {
    const guard = new SubmitGuard();
    const src = new Subject<number>();
    const seen: number[] = [];

    guard.run(src, { next: (v) => seen.push(v) });
    expect(guard.pending()).toBe(true);

    src.next(1);
    src.complete();
    expect(seen).toEqual([1]);
    expect(guard.pending()).toBe(false);
  });

  it('ignores a re-entrant call while pending (no second subscribe)', () => {
    const guard = new SubmitGuard();
    const first = new Subject<void>();
    guard.run(first, {});

    // A double-click: second call arrives before the first resolves.
    let secondSubscribed = false;
    guard.run(
      new Observable<void>(() => {
        secondSubscribed = true;
      }),
      {},
    );
    expect(secondSubscribed).toBe(false);

    // Once the first resolves, a fresh call runs again.
    first.complete();
    expect(guard.pending()).toBe(false);
    let thirdSubscribed = false;
    guard.run(
      new Observable<void>(() => {
        thirdSubscribed = true;
      }),
      {},
    );
    expect(thirdSubscribed).toBe(true);
  });

  it('re-enables after an error so the user can retry', () => {
    const guard = new SubmitGuard();
    const src = new Subject<void>();
    let errored = false;

    guard.run(src, {
      error: () => {
        errored = true;
      },
    });
    expect(guard.pending()).toBe(true);

    src.error(new Error('boom'));
    expect(errored).toBe(true);
    expect(guard.pending()).toBe(false);
  });
});
