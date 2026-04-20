import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { Customer } from '../../models/customer';
import { CustomerService } from '../../services/customer.service';
import { CustomerCreateModalComponent } from '../customer-create/customer-create-modal.component';

@Component({
  selector: 'app-customer-list',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, CustomerCreateModalComponent],
  templateUrl: './customer-list.component.html',
  styleUrl: './customer-list.component.scss',
})
export class CustomerListComponent {
  private readonly customersApi = inject(CustomerService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly customers = signal<Customer[]>([]);
  readonly showCreateModal = signal(false);

  constructor() {
    this.customersApi.list(100).subscribe({
      next: (items) => {
        this.customers.set(items ?? []);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message ?? 'Failed to load customers');
        this.loading.set(false);
      },
    });
  }
}
