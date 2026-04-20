import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CreateProductRequest, Product } from '../models/product';

@Injectable({ providedIn: 'root' })
export class ProductService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<Product[]> {
    const params = new HttpParams().set('limit', String(limit));
    return this.http.get<Product[]>(`${environment.apiBaseUrl}/products`, { params });
  }

  getById(id: string): Observable<Product> {
    return this.http.get<Product>(`${environment.apiBaseUrl}/products/${encodeURIComponent(id)}`);
  }

  create(body: CreateProductRequest): Observable<Product> {
    return this.http.post<Product>(`${environment.apiBaseUrl}/products`, body);
  }
}
