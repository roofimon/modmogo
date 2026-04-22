import { HttpErrorResponse } from '@angular/common/http';
import * as E from 'fp-ts/Either';
import { Observable } from 'rxjs';
import { catchError, map, of } from 'rxjs';

export interface ApiError {
  readonly status: number;
  readonly message: string;
}

function extractMessage(err: HttpErrorResponse): string {
  if (typeof err.error?.message === 'string') return err.error.message;
  if (typeof err.error === 'string') return err.error;
  return err.message;
}

export function fromHttp<T>(obs$: Observable<T>): Observable<E.Either<ApiError, T>> {
  return obs$.pipe(
    map(E.right),
    catchError((err: unknown) =>
      of(
        E.left<ApiError>({
          status: err instanceof HttpErrorResponse ? err.status : 0,
          message:
            err instanceof HttpErrorResponse
              ? extractMessage(err)
              : err instanceof Error
                ? err.message
                : 'Unknown error',
        }),
      ),
    ),
  );
}
