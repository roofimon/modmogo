import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { CustomerService } from '../../services/customer.service';

@Component({
  selector: 'app-customer-create',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './customer-create.component.html',
  styleUrl: './customer-create.component.scss',
})
export class CustomerCreateComponent {
  private readonly fb = inject(FormBuilder);
  private readonly customersApi = inject(CustomerService);
  private readonly router = inject(Router);

  readonly submitting = signal(false);
  readonly error = signal<string | null>(null);

  readonly form = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(500)]],
    email: ['', [Validators.required, Validators.maxLength(320), Validators.email]],
  });

  submit(): void {
    this.error.set(null);
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const v = this.form.getRawValue();
    this.submitting.set(true);
    this.customersApi
      .create({
        name: v.name.trim(),
        email: v.email.trim(),
      })
      .subscribe({
        next: (c) => {
          void this.router.navigate(['/customers', c.id]);
        },
        error: (e: unknown) => {
          const msg =
            e instanceof HttpErrorResponse
              ? typeof e.error === 'string'
                ? e.error
                : e.message
              : e instanceof Error
                ? e.message
                : 'Could not create customer';
          this.error.set(msg);
          this.submitting.set(false);
        },
      });
  }
}
