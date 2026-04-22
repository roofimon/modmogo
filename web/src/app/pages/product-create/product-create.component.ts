import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

import { ProductService } from '../../services/product.service';

@Component({
  selector: 'app-product-create',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './product-create.component.html',
  styleUrl: './product-create.component.scss',
})
export class ProductCreateComponent {
  private readonly fb = inject(FormBuilder);
  private readonly productsApi = inject(ProductService);
  private readonly router = inject(Router);

  readonly submitting = signal(false);
  readonly error = signal<string | null>(null);

  readonly form = this.fb.nonNullable.group({
    sku: ['', [Validators.required, Validators.maxLength(64)]],
    name: ['', [Validators.required, Validators.maxLength(500)]],
    price: [0, [Validators.required, Validators.min(0)]],
  });

  submit(): void {
    this.error.set(null);
    if (this.form.invalid) { this.form.markAllAsTouched(); return; }
    const v = this.form.getRawValue();
    this.submitting.set(true);
    this.productsApi.create({ sku: v.sku.trim(), name: v.name.trim(), price: v.price }).subscribe((result) =>
      pipe(
        result,
        E.fold(
          (err) => { this.error.set(err.message); this.submitting.set(false); },
          (p) => { void this.router.navigate(['/products', p.id]); },
        ),
      ),
    );
  }
}
