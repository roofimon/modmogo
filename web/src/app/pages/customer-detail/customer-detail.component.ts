import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { switchMap, throwError } from 'rxjs';

import { Customer } from '../../models/customer';
import { CustomerService } from '../../services/customer.service';

@Component({
  selector: 'app-customer-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './customer-detail.component.html',
  styleUrl: './customer-detail.component.scss',
})
export class CustomerDetailComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly customersApi = inject(CustomerService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly customer = signal<Customer | null>(null);
  readonly deactivateSubmitting = signal(false);
  readonly deactivateError = signal<string | null>(null);

  constructor() {
    this.route.paramMap
      .pipe(
        switchMap((params) => {
          const id = params.get('id');
          if (!id) {
            return throwError(() => new Error('Missing customer id'));
          }
          return this.customersApi.getById(id);
        }),
      )
      .subscribe({
        next: (c) => {
          this.customer.set(c);
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
                : 'Failed to load customer';
          this.error.set(msg);
          this.loading.set(false);
        },
      });
  }

  isInactive(c: Customer): boolean {
    return c.deactivated_at != null && c.deactivated_at !== '';
  }

  confirmDeactivate(): void {
    if (!confirm('Deactivate this customer? They will disappear from the directory.')) {
      return;
    }
    const c = this.customer();
    if (!c) return;
    this.deactivateError.set(null);
    this.deactivateSubmitting.set(true);
    this.customersApi.deactivate(c.id).subscribe({
      next: (updated) => {
        this.customer.set(updated);
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
