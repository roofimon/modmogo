import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CreateProductRequest, Product } from '../models/product';
import { ApiError, fromHttp } from './api-error';

@Injectable({ providedIn: 'root' })
export class ProductService {
  private readonly http = inject(HttpClient);

  list(limit = 50): Observable<E.Either<ApiError, Product[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Product[]>(`${environment.apiBaseUrl}/products`, { params }), fromHttp);
  }

  listInactive(limit = 50): Observable<E.Either<ApiError, Product[]>> {
    const params = new HttpParams().set('limit', String(limit));
    return pipe(this.http.get<Product[]>(`${environment.apiBaseUrl}/products/inactive`, { params }), fromHttp);
  }

  getById(id: string): Observable<E.Either<ApiError, Product>> {
    return pipe(this.http.get<Product>(`${environment.apiBaseUrl}/products/${encodeURIComponent(id)}`), fromHttp);
  }

  create(body: CreateProductRequest): Observable<E.Either<ApiError, Product>> {
    return pipe(this.http.post<Product>(`${environment.apiBaseUrl}/products`, body), fromHttp);
  }

  deactivate(id: string): Observable<E.Either<ApiError, Product>> {
    return pipe(
      this.http.post<Product>(`${environment.apiBaseUrl}/products/${encodeURIComponent(id)}/deactivate`, {}),
      fromHttp,
    );
  }

  activate(id: string): Observable<E.Either<ApiError, Product>> {
    return pipe(
      this.http.post<Product>(`${environment.apiBaseUrl}/products/${encodeURIComponent(id)}/activate`, {}),
      fromHttp,
    );
  }
}
