import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';
import { of, switchMap } from 'rxjs';

import { Customer } from '../../models/customer';
import { ApiError } from '../../services/api-error';
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
          if (!id) return of(E.left<ApiError, Customer>({ status: 0, message: 'Missing customer id' }));
          return this.customersApi.getById(id);
        }),
      )
      .subscribe((result) =>
        pipe(
          result,
          E.fold(
            (err) => { this.error.set(err.message); this.loading.set(false); },
            (c) => { this.customer.set(c); this.loading.set(false); },
          ),
        ),
      );
  }

  isInactive(c: Customer): boolean {
    return c.deactivated_at != null && c.deactivated_at !== '';
  }

  confirmDeactivate(): void {
    if (!confirm('Deactivate this customer? They will disappear from the directory.')) return;
    const c = this.customer();
    if (!c) return;
    this.deactivateError.set(null);
    this.deactivateSubmitting.set(true);
    this.customersApi.deactivate(c.id).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.deactivateError.set(err.message); this.deactivateSubmitting.set(false); },
          (updated) => { this.customer.set(updated); this.deactivateSubmitting.set(false); },
        ),
      ),
    );
  }
}
