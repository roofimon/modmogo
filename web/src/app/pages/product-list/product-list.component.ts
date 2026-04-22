import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

import { Product } from '../../models/product';
import { ProductService } from '../../services/product.service';
import { ProductCreateModalComponent } from '../product-create/product-create-modal.component';

@Component({
  selector: 'app-product-list',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, ProductCreateModalComponent],
  templateUrl: './product-list.component.html',
  styleUrl: './product-list.component.scss',
})
export class ProductListComponent {
  private readonly productsApi = inject(ProductService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly products = signal<Product[]>([]);
  readonly showCreateModal = signal(false);

  constructor() {
    this.productsApi.list(100).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.error.set(err.message); this.loading.set(false); },
          (items) => { this.products.set(items ?? []); this.loading.set(false); },
        ),
      ),
    );
  }
}
