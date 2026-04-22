import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';
import { of, switchMap } from 'rxjs';

import { Product } from '../../models/product';
import { ApiError } from '../../services/api-error';
import { ProductService } from '../../services/product.service';

@Component({
  selector: 'app-product-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './product-detail.component.html',
  styleUrl: './product-detail.component.scss',
})
export class ProductDetailComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly productsApi = inject(ProductService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly product = signal<Product | null>(null);
  readonly deactivateSubmitting = signal(false);
  readonly deactivateError = signal<string | null>(null);
  readonly activateSubmitting = signal(false);
  readonly activateError = signal<string | null>(null);

  constructor() {
    this.route.paramMap
      .pipe(
        switchMap((params) => {
          const id = params.get('id');
          if (!id) return of(E.left<ApiError, Product>({ status: 0, message: 'Missing product id' }));
          return this.productsApi.getById(id);
        }),
      )
      .subscribe((result) =>
        pipe(
          result,
          E.fold(
            (err) => { this.error.set(err.message); this.loading.set(false); },
            (p) => { this.product.set(p); this.loading.set(false); },
          ),
        ),
      );
  }

  isInactive(p: Product): boolean {
    return p.deactivated_at != null && p.deactivated_at !== '';
  }

  confirmActivate(): void {
    if (!confirm('Activate this product? It will reappear in the catalog.')) return;
    const p = this.product();
    if (!p) return;
    this.activateError.set(null);
    this.activateSubmitting.set(true);
    this.productsApi.activate(p.id).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.activateError.set(err.message); this.activateSubmitting.set(false); },
          (updated) => { this.product.set(updated); this.activateSubmitting.set(false); },
        ),
      ),
    );
  }

  confirmDeactivate(): void {
    if (!confirm('Deactivate this product? It will disappear from the catalog.')) return;
    const p = this.product();
    if (!p) return;
    this.deactivateError.set(null);
    this.deactivateSubmitting.set(true);
    this.productsApi.deactivate(p.id).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.deactivateError.set(err.message); this.deactivateSubmitting.set(false); },
          (updated) => { this.product.set(updated); this.deactivateSubmitting.set(false); },
        ),
      ),
    );
  }
}
