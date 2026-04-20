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

### Modals
- Add product → `ProductCreateModalComponent` (opened from product list pages)
- Add customer → `CustomerCreateModalComponent` (opened from customer list pages)
