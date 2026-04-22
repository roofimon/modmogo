import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Output, inject, signal } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import * as E from 'fp-ts/Either';
import { pipe } from 'fp-ts/function';

import { CatalogCustomer, CatalogProduct } from '../../models/order';
import { OrderService } from '../../services/order.service';

@Component({
  selector: 'app-order-create-modal',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './order-create-modal.component.html',
  styleUrl: './order-create-modal.component.scss',
})
export class OrderCreateModalComponent {
  @Output() closed = new EventEmitter<void>();

  private readonly fb = inject(FormBuilder);
  private readonly ordersApi = inject(OrderService);
  private readonly router = inject(Router);

  readonly submitting = signal(false);
  readonly error = signal<string | null>(null);

  // SKU autocomplete
  readonly products = signal<CatalogProduct[]>([]);
  readonly activeSuggestionRow = signal<number | null>(null);
  readonly suggestions = signal<CatalogProduct[]>([]);

  // Customer autocomplete
  readonly allCustomers = signal<CatalogCustomer[]>([]);
  readonly customerSearchText = signal('');
  readonly customerSuggestions = signal<CatalogCustomer[]>([]);
  readonly showCustomerDropdown = signal(false);

  readonly form = this.fb.nonNullable.group({
    customer_id: ['', Validators.pattern(/^$|^[0-9a-fA-F]{24}$/)],
    items: this.fb.array([this.buildItemGroup()]),
  });

  constructor() {
    this.ordersApi.listProducts().subscribe((result) =>
      pipe(result, E.fold(() => {}, (items) => this.products.set(items ?? []))),
    );
    this.ordersApi.listCustomers().subscribe((result) =>
      pipe(result, E.fold(() => {}, (items) => this.allCustomers.set(items ?? []))),
    );
  }

  get items(): FormArray {
    return this.form.controls.items as FormArray;
  }

  itemAt(i: number): FormGroup {
    return this.items.at(i) as FormGroup;
  }

  private buildItemGroup(): FormGroup {
    return this.fb.nonNullable.group({
      sku:        ['', [Validators.required, Validators.maxLength(64)]],
      quantity:   [1,  [Validators.required, Validators.min(1)]],
      unit_price: [0,  [Validators.required, Validators.min(0)]],
    });
  }

  addItem(): void {
    this.items.push(this.buildItemGroup());
  }

  removeItem(index: number): void {
    if (this.items.length > 1) {
      this.items.removeAt(index);
    }
  }

  // Customer autocomplete
  onCustomerInput(value: string): void {
    this.customerSearchText.set(value);
    const q = value.trim().toLowerCase();
    if (!q) {
      this.form.controls.customer_id.setValue('');
      this.customerSuggestions.set([]);
      this.showCustomerDropdown.set(false);
      return;
    }
    this.customerSuggestions.set(
      this.allCustomers().filter((c) => c.name.toLowerCase().startsWith(q)).slice(0, 8),
    );
    this.showCustomerDropdown.set(true);
  }

  selectCustomer(c: CatalogCustomer): void {
    this.form.controls.customer_id.setValue(c.id);
    this.customerSearchText.set(c.name);
    this.customerSuggestions.set([]);
    this.showCustomerDropdown.set(false);
  }

  clearCustomerSuggestions(): void {
    setTimeout(() => this.showCustomerDropdown.set(false), 150);
  }

  // SKU autocomplete
  onSkuInput(value: string, i: number): void {
    const q = value.trim().toLowerCase();
    if (!q) {
      this.suggestions.set([]);
      this.activeSuggestionRow.set(null);
      return;
    }
    this.activeSuggestionRow.set(i);
    this.suggestions.set(
      this.products().filter((p) => p.sku.toLowerCase().startsWith(q)).slice(0, 8),
    );
  }

  selectSku(product: CatalogProduct, i: number): void {
    const row = this.itemAt(i);
    row.controls['sku'].setValue(product.sku);
    row.controls['unit_price'].setValue(product.price);
    this.activeSuggestionRow.set(null);
    this.suggestions.set([]);
  }

  clearSuggestions(): void {
    setTimeout(() => {
      this.activeSuggestionRow.set(null);
      this.suggestions.set([]);
    }, 150);
  }

  close(): void {
    this.closed.emit();
  }

  submit(): void {
    this.error.set(null);
    if (this.form.invalid) { this.form.markAllAsTouched(); return; }
    const v = this.form.getRawValue();
    const customerId = v.customer_id.trim();
    this.submitting.set(true);
    this.ordersApi
      .create({
        customer_id: customerId || null,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        items: v.items.map((i: any) => ({
          sku: (i['sku'] as string).trim(),
          quantity: i['quantity'] as number,
          unit_price: i['unit_price'] as number,
        })),
      })
      .subscribe((result) =>
        pipe(
          result,
          E.fold(
            (err) => { this.error.set(err.message); this.submitting.set(false); },
            (o) => { void this.router.navigate(['/orders', o.id]); },
          ),
        ),
      );
  }
}
