import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

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
    if (this.form.invalid) { this.form.markAllAsTouched(); return; }
    const v = this.form.getRawValue();
    this.submitting.set(true);
    this.customersApi.create({ name: v.name.trim(), email: v.email.trim() }).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.error.set(err.message); this.submitting.set(false); },
          (c) => { void this.router.navigate(['/customers', c.id]); },
        ),
      ),
    );
  }
}
