import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

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
    this.productsApi.listInactive(100).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.error.set(err.message); this.loading.set(false); },
          (items) => { this.products.set(items ?? []); this.loading.set(false); },
        ),
      ),
    );
  }

  activate(id: string, event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    this.activateError.set(null);
    this.activating.set(id);
    this.productsApi.activate(id).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.activateError.set(err.message); this.activating.set(null); },
          () => { this.products.update((list) => list.filter((p) => p.id !== id)); this.activating.set(null); },
        ),
      ),
    );
  }
}
