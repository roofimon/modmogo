import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CatalogCustomer, CatalogProduct, CreateOrderRequest, Order } from '../models/order';

@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<Order[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Order[]>(`${environment.apiBaseUrl}/orders`, { params });
  }

  listInactive(limit = 50): Observable<Order[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Order[]>(`${environment.apiBaseUrl}/orders/inactive`, { params });
  }

  getById(id: string): Observable<Order> {
    return this.http.get<Order>(`${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}`);
  }

  create(body: CreateOrderRequest): Observable<Order> {
    return this.http.post<Order>(`${environment.apiBaseUrl}/orders`, body);
  }

  listCustomers(limit = 100): Observable<CatalogCustomer[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<CatalogCustomer[]>(`${environment.apiBaseUrl}/orders/customers`, { params });
  }

  listProducts(limit = 100): Observable<CatalogProduct[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<CatalogProduct[]>(`${environment.apiBaseUrl}/orders/products`, { params });
  }

  deactivate(id: string): Observable<Order> {
    return this.http.post<Order>(
      `${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}/deactivate`,
      {},
    );
  }

  listPaymentCompleted(limit = 100): Observable<Order[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Order[]>(`${environment.apiBaseUrl}/orders/payment-completed`, { params });
  }

  completePayment(id: string): Observable<Order> {
    return this.http.post<Order>(
      `${environment.apiBaseUrl}/orders/${encodeURIComponent(id)}/complete-payment`,
      {},
    );
  }
}
