import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { switchMap, throwError } from 'rxjs';

import { Order } from '../../models/order';
import { OrderService } from '../../services/order.service';

@Component({
  selector: 'app-order-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './order-detail.component.html',
  styleUrl: './order-detail.component.scss',
})
export class OrderDetailComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly ordersApi = inject(OrderService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly order = signal<Order | null>(null);
  readonly deactivateSubmitting = signal(false);
  readonly deactivateError = signal<string | null>(null);
  readonly paymentSubmitting = signal(false);
  readonly paymentError = signal<string | null>(null);

  constructor() {
    this.route.paramMap
      .pipe(
        switchMap((params) => {
          const id = params.get('id');
          if (!id) {
            return throwError(() => new Error('Missing order id'));
          }
          return this.ordersApi.getById(id);
        }),
      )
      .subscribe({
        next: (o) => {
          this.order.set(o);
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
                : 'Failed to load order';
          this.error.set(msg);
          this.loading.set(false);
        },
      });
  }

  isInactive(o: Order): boolean {
    return o.deactivated_at != null && o.deactivated_at !== '';
  }

  isPaymentCompleted(o: Order): boolean {
    return o.status === 'payment_completed';
  }

  subtotal(quantity: number, unit_price: number): number {
    return quantity * unit_price;
  }

  completePayment(): void {
    const o = this.order();
    if (!o || this.isInactive(o) || this.isPaymentCompleted(o)) return;
    this.paymentError.set(null);
    this.paymentSubmitting.set(true);
    this.ordersApi.completePayment(o.id).subscribe({
      next: (newOrder) => {
        this.paymentSubmitting.set(false);
        void this.router.navigate(['/orders', newOrder.id]);
      },
      error: (e: unknown) => {
        const msg =
          e instanceof HttpErrorResponse
            ? (e.error?.message ?? e.message)
            : e instanceof Error
              ? e.message
              : 'Could not complete payment';
        this.paymentError.set(msg);
        this.paymentSubmitting.set(false);
      },
    });
  }

  confirmDeactivate(): void {
    if (!confirm('Deactivate this order? It will move to the inactive list.')) {
      return;
    }
    const o = this.order();
    if (!o) return;
    this.deactivateError.set(null);
    this.deactivateSubmitting.set(true);
    this.ordersApi.deactivate(o.id).subscribe({
      next: (updated) => {
        this.order.set(updated);
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
