import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

import { Customer } from '../../models/customer';
import { CustomerService } from '../../services/customer.service';
import { CustomerCreateModalComponent } from '../customer-create/customer-create-modal.component';

@Component({
  selector: 'app-customer-inactive-list',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, CustomerCreateModalComponent],
  templateUrl: './customer-inactive-list.component.html',
  styleUrl: './customer-inactive-list.component.scss',
})
export class CustomerInactiveListComponent {
  private readonly customersApi = inject(CustomerService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly customers = signal<Customer[]>([]);
  readonly showCreateModal = signal(false);

  constructor() {
    this.customersApi.listInactive(100).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.error.set(err.message); this.loading.set(false); },
          (items) => { this.customers.set(items ?? []); this.loading.set(false); },
        ),
      ),
    );
  }
}
