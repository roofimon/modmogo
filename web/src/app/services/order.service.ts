import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CatalogCustomer, CatalogProduct, CreateOrderRequest, Order } from '../models/order';
import { ApiError, fromHttp } from './api-error';

@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<E.Either<ApiError, Order[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Order[]>(`${environment.apiBaseUrl}/orders`, { params }), fromHttp);
  }

  listInactive(limit = 50): Observable<E.Either<ApiError, Order[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Order[]>(`${environment.apiBaseUrl}/orders/inactive`, { params }), fromHttp);
  }

  listPaymentCompleted(limit = 50): Observable<E.Either<ApiError, Order[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Order[]>(`${environment.apiBaseUrl}/orders/payment-completed`, { params }), fromHttp);
  }

  getById(id: string): Observable<E.Either<ApiError, Order>> {
    return pipe(this.http.get<Order>(`${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}`), fromHttp);
  }

  create(body: CreateOrderRequest): Observable<E.Either<ApiError, Order>> {
    return pipe(this.http.post<Order>(`${environment.apiBaseUrl}/orders`, body), fromHttp);
  }

  listProducts(limit = 100): Observable<E.Either<ApiError, CatalogProduct[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<CatalogProduct[]>(`${environment.apiBaseUrl}/orders/products`, { params }), fromHttp);
  }

  listCustomers(limit = 100): Observable<E.Either<ApiError, CatalogCustomer[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<CatalogCustomer[]>(`${environment.apiBaseUrl}/orders/customers`, { params }), fromHttp);
  }

  deactivate(id: string): Observable<E.Either<ApiError, Order>> {
    return pipe(
      this.http.post<Order>(`${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}/deactivate`, {}),
      fromHttp,
    );
  }

  completePayment(id: string): Observable<E.Either<ApiError, Order>> {
    return pipe(
      this.http.post<Order>(`${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}/complete-payment`, {}),
      fromHttp,
    );
  }
}
