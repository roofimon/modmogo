import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { Product } from '../../models/product';
import { ProductService } from '../../services/product.service';
import { ProductCreateModalComponent } from '../product-create/product-create-modal.component';

@Component({
  selector: 'app-product-inactive-list',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, ProductCreateModalComponent],
  templateUrl: './product-inactive-list.component.html',
  styleUrl: './product-inactive-list.component.scss',
})
export class ProductInactiveListComponent {
  private readonly productsApi = inject(ProductService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly products = signal<Product[]>([]);
  readonly showCreateModal = signal(false);
  readonly activating = signal<string | null>(null);
  readonly activateError = signal<string | null>(null);

  constructor() {
    this.productsApi.listInactive(100).subscribe({
      next: (items) => {
        this.products.set(items ?? []);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message ?? 'Failed to load inactive products');
        this.loading.set(false);
      },
    });
  }

  activate(id: string, event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    this.activateError.set(null);
    this.activating.set(id);
    this.productsApi.activate(id).subscribe({
      next: () => {
        this.products.update((list) => list.filter((p) => p.id !== id));
        this.activating.set(null);
      },
      error: (e: unknown) => {
        const msg =
          e instanceof HttpErrorResponse
            ? typeof e.error === 'string'
              ? e.error
              : e.message
            : e instanceof Error
              ? e.message
              : 'Could not activate';
        this.activateError.set(msg);
        this.activating.set(null);
      },
    });
  }
}
