# Architecture

## Back End
- Modular Monolith — each domain lives under `internal/<domain>/`
- Ports and Adapters pattern — handler → service → repository interface → MongoDB implementation
- NO shared services between domains
- `samber/mo` for Result/Option types
- Functional style: small pure functions, side-effect functions, and orchestration functions kept separate

### Domains
| Domain | Routes |
|--------|--------|
| product | `GET/POST /products`, `GET /products/inactive`, `GET/POST /products/:id`, `POST /products/:id/deactivate`, `POST /products/:id/activate` |
| customer | `GET/POST /customers`, `GET /customers/inactive`, `GET/POST /customers/:id`, `POST /customers/:id/deactivate` |
| order | `GET/POST /orders`, `GET /orders/inactive`, `GET /orders/products`, `GET /orders/:id`, `POST /orders/:id/deactivate` |

## Front End
- Angular 19 standalone components, signals for state
- Small components; lazy-loaded routes
- Coinbase-inspired design system (`web/src/_coinbase-tokens.scss`)
- Top sticky dark nav; pill buttons; Active/Inactive tab bars per section

### Pages
| Route | Component |
|-------|-----------|
| `/products` | ProductListComponent |
| `/products/inactive` | ProductInactiveListComponent |
| `/products/:id` | ProductDetailComponent |
| `/customers` | CustomerListComponent |
| `/customers/inactive` | CustomerInactiveListComponent |
| `/customers/:id` | CustomerDetailComponent |
| `/orders` | OrderListComponent |
| `/orders/inactive` | OrderInactiveListComponent |
| `/orders/:id` | OrderDetailComponent |

### Modals
- Add product → `ProductCreateModalComponent` (opened from product list pages)
- Add customer → `CustomerCreateModalComponent` (opened from customer list pages)
- Add order → `OrderCreateModalComponent` (opened from order list pages; includes SKU autocomplete)

# Product Requirement

## Product
- Browse active products in a card grid; switch to Inactive tab to see soft-deleted products.
- Create a product via modal (SKU, name, price). SKU must be unique.
- View product detail: SKU, name, price, created date, deactivated date (if any).
- Deactivate an active product (soft-delete — moves to Inactive list).
- Activate an inactive product (restores it to the active catalog).

## Customer
- Browse active customers in a card grid; switch to Inactive tab for deactivated customers.
- Create a customer via modal (name, email).
- View customer detail: name, email, created date, deactivated date (if any).
- Deactivate a customer (soft-delete).
- Add mobile phone number field

## Order
- Browse active orders in a card grid showing item count, optional customer ID, and total; switch to Inactive tab for deactivated orders.
- Create an order via modal:
  - Optional customer ID field (must be a valid 24-char hex ObjectID or left blank).
  - Dynamic line-item rows (SKU, quantity, unit price); at least one item required.
  - SKU field has autocomplete: suggests active products by prefix as the user types, auto-fills unit price on selection.
  - Line items can be added or removed; minimum one row enforced.
- View order detail: line-items table (SKU / Qty / Unit Price / Subtotal) with a Total footer row; customer ID links to the customer detail page.
- Deactivate an order (soft-delete — moves to Inactive list). No activate action.
- Order total is computed at read time (never stored); the product catalog is fetched via `GET /orders/products` — the order domain defines its own `ProductCatalog` port so it does not import the product domain directly.
- In Order detail it should display customer name and product name, so GetByID must query product and customer name from database. We must create order display model that contains 2 additional field. 
- In Order detail it should have Payment Completed button. Once payment completed order service will create new Order which contains ID from original order but this one status marked as "PaymentCompleted"