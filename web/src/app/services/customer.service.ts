import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CreateCustomerRequest, Customer } from '../models/customer';
import { ApiError, fromHttp } from './api-error';

@Injectable({ providedIn: 'root' })
export class CustomerService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<E.Either<ApiError, Customer[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Customer[]>(`${environment.apiBaseUrl}/customers`, { params }), fromHttp);
  }

  listInactive(limit = 50): Observable<E.Either<ApiError, Customer[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Customer[]>(`${environment.apiBaseUrl}/customers/inactive`, { params }), fromHttp);
  }

  getById(id: string): Observable<E.Either<ApiError, Customer>> {
    return pipe(this.http.get<Customer>(`${environment.apiBaseUrl}/customers/${encodeURIComponent(id)}`), fromHttp);
  }

  create(body: CreateCustomerRequest): Observable<E.Either<ApiError, Customer>> {
    return pipe(this.http.post<Customer>(`${environment.apiBaseUrl}/customers`, body), fromHttp);
  }

  deactivate(id: string): Observable<E.Either<ApiError, Customer>> {
    return pipe(
      this.http.post<Customer>(`${environment.apiBaseUrl}/customers/${encodeURIComponent(id)}/deactivate`, {}),
      fromHttp,
    );
  }
}
