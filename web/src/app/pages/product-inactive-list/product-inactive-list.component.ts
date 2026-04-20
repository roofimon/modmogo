import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { Product } from '../../models/product';
import { ProductService } from '../../services/product.service';

@Component({
  selector: 'app-product-inactive-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './product-inactive-list.component.html',
  styleUrl: './product-inactive-list.component.scss',
})
export class ProductInactiveListComponent {
  private readonly productsApi = inject(ProductService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly products = signal<Product[]>([]);

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
}
