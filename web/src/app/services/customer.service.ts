import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CreateCustomerRequest, Customer } from '../models/customer';

@Injectable({ providedIn: 'root' })
export class CustomerService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<Customer[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Customer[]>(`${environment.apiBaseUrl}/customers`, { params });
  }

  /** Deactivated customers only (newest deactivation first). */
  listInactive(limit = 50): Observable<Customer[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Customer[]>(`${environment.apiBaseUrl}/customers/inactive`, { params });
  }

  getById(id: string): Observable<Customer> {
    return this.http.get<Customer>(`${environment.apiBaseUrl}/customers/${encodeURIComponent(id)}`);
  }

  create(body: CreateCustomerRequest): Observable<Customer> {
    return this.http.post<Customer>(`${environment.apiBaseUrl}/customers`, body);
  }

  deactivate(id: string): Observable<Customer> {
    return this.http.post<Customer>(
      `${environment.apiBaseUrl}/customers/${encodeURIComponent(id)}/deactivate`,
      {},
    );
  }
}
