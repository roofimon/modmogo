import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { Order } from '../../models/order';
import { OrderService } from '../../services/order.service';
import { OrderCreateModalComponent } from '../order-create/order-create-modal.component';

@Component({
  selector: 'app-order-inactive-list',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, OrderCreateModalComponent],
  templateUrl: './order-inactive-list.component.html',
  styleUrl: './order-inactive-list.component.scss',
})
export class OrderInactiveListComponent {
  private readonly ordersApi = inject(OrderService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly orders = signal<Order[]>([]);
  readonly showCreateModal = signal(false);

  constructor() {
    this.ordersApi.listInactive(100).subscribe({
      next: (items) => {
        this.orders.set(items ?? []);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message ?? 'Failed to load inactive orders');
        this.loading.set(false);
      },
    });
  }
}
