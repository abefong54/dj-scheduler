import { inject } from '@angular/core';
import { HttpInterceptorFn } from '@angular/common/http';
import { finalize } from 'rxjs';
import { LoadingService } from '../services/loading.service';

/**
 * Marks a request in-flight with LoadingService for the duration of the call so
 * the global progress bar reflects any activity — reads and mutations alike.
 * finalize() runs on success, error, and cancellation, so the counter always
 * balances.
 */
export const loadingInterceptor: HttpInterceptorFn = (req, next) => {
  const loading = inject(LoadingService);
  loading.start();
  return next(req).pipe(finalize(() => loading.stop()));
};
