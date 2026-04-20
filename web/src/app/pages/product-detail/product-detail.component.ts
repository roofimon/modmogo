import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { switchMap, throwError } from 'rxjs';

import { Product } from '../../models/product';
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

  constructor() {
    this.route.paramMap
      .pipe(
        switchMap((params) => {
          const id = params.get('id');
          if (!id) {
            return throwError(() => new Error('Missing product id'));
          }
          return this.productsApi.getById(id);
        }),
      )
      .subscribe({
        next: (p) => {
          this.product.set(p);
          this.loading.set(false);
        },
        error: (e: unknown) => {
          const msg =
            e instanceof HttpErrorResponse
              ? typeof e.error === 'string'
                ? e.error
                : e.message
              : e instanceof Error
                ? e.message
                : 'Failed to load product';
          this.error.set(msg);
          this.loading.set(false);
        },
      });
  }

  isInactive(p: Product): boolean {
    return p.deactivated_at != null && p.deactivated_at !== '';
  }

  confirmDeactivate(): void {
    if (!confirm('Deactivate this product? It will disappear from the catalog.')) {
      return;
    }
    const p = this.product();
    if (!p) return;
    this.deactivateError.set(null);
    this.deactivateSubmitting.set(true);
    this.productsApi.deactivate(p.id).subscribe({
      next: (updated) => {
        this.product.set(updated);
        this.deactivateSubmitting.set(false);
      },
      error: (e: unknown) => {
        const msg =
          e instanceof HttpErrorResponse
            ? typeof e.error === 'string'
              ? e.error
              : e.message
            : e instanceof Error
              ? e.message
              : 'Could not deactivate';
        this.deactivateError.set(msg);
        this.deactivateSubmitting.set(false);
      },
    });
  }
}
