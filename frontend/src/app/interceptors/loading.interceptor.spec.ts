import { TestBed } from '@angular/core/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { loadingInterceptor } from './loading.interceptor';
import { LoadingService } from '../services/loading.service';

describe('loadingInterceptor', () => {
  let http: HttpClient;
  let httpCtrl: HttpTestingController;
  let loading: LoadingService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([loadingInterceptor])),
        provideHttpClientTesting(),
      ],
    });
    http = TestBed.inject(HttpClient);
    httpCtrl = TestBed.inject(HttpTestingController);
    loading = TestBed.inject(LoadingService);
  });
  afterEach(() => httpCtrl.verify());

  it('is loading while the request is open, idle once it resolves', () => {
    expect(loading.isLoading()).toBe(false);
    http.get('/api/events').subscribe();
    expect(loading.isLoading()).toBe(true);
    httpCtrl.expectOne('/api/events').flush([]);
    expect(loading.isLoading()).toBe(false);
  });

  it('balances the counter even when the request errors', () => {
    http.get('/api/events').subscribe({ error: () => {} });
    expect(loading.isLoading()).toBe(true);
    httpCtrl.expectOne('/api/events').flush(null, { status: 500, statusText: 'err' });
    expect(loading.isLoading()).toBe(false);
  });
});
