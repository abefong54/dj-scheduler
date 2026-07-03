import { TestBed } from '@angular/core/testing';
import { LoadingService } from './loading.service';

describe('LoadingService', () => {
  let svc: LoadingService;
  beforeEach(() => {
    TestBed.configureTestingModule({});
    svc = TestBed.inject(LoadingService);
  });

  it('is idle initially', () => {
    expect(svc.isLoading()).toBe(false);
  });

  it('is loading while a request is active', () => {
    svc.start();
    expect(svc.isLoading()).toBe(true);
  });

  it('stays loading until every concurrent request finishes', () => {
    svc.start();
    svc.start();
    svc.stop();
    expect(svc.isLoading()).toBe(true);
    svc.stop();
    expect(svc.isLoading()).toBe(false);
  });

  it('never underflows below idle', () => {
    svc.stop();
    expect(svc.isLoading()).toBe(false);
    svc.start();
    expect(svc.isLoading()).toBe(true);
  });
});
