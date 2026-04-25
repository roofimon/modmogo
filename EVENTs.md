# Event Inventory

## 1. Domain Events

Things that happened, in business language. Emitted after a successful state change.

+------------------------------------+------------------+-----------------------------------------------+
| Event                              | Status           | Trigger                                       |
+------------------------------------+------------------+-----------------------------------------------+
| product.created                    | [x] implemented  | Create                                        |
| product.activated                  | [x] implemented  | Activate                                      |
| product.deactivated                | [x] implemented  | Deactivate                                    |
| customer.registered                | [x] implemented  | Create                                        |
| customer.deactivated               | [x] implemented  | Deactivate                                    |
| order.placed                       | [x] implemented  | PlaceOrder                                    |
| order.payment_completed            | [x] implemented  | CompletePayment                               |
| order.cancelled                    | [x] implemented  | Deactivate                                    |
| product.create_rejected            | [ ] missing      | validation failure in Create                  |
| customer.register_rejected         | [ ] missing      | validation failure in Create                  |
| order.place_rejected               | [ ] missing      | validation failure in PlaceOrder              |
| order.payment_already_completed    | [ ] missing      | ErrAlreadyCompleted in CompletePayment        |
+------------------------------------+------------------+-----------------------------------------------+

## 2. Event Sourcing Events

State changes recorded as immutable facts — the log *is* the state.

+-------------------------+------------------+----------------------------------------------------------+
| Event                   | Status           | Notes                                                    |
+-------------------------+------------------+----------------------------------------------------------+
| order.payment_completed | [~] partial      | CompletePayment appends a new Order record with          |
|                         |                  | OriginalOrderID — the only append-only operation         |
| product.deactivated     | [ ] mutates      | currently updates status + deactivated_at in place       |
| product.activated       | [ ] mutates      | clears deactivated_at in place                           |
| customer.deactivated    | [ ] mutates      | updates status + deactivated_at in place                 |
| order.cancelled         | [ ] mutates      | updates deactivated_at in place                          |
+-------------------------+------------------+----------------------------------------------------------+

Full event sourcing would mean storing these as new records (like CompletePayment does) instead of mutating existing ones.

## 3. Integration Events

Cross-boundary signals — what external systems or other services would care about.

+-------------------------+---------------------------------------------------+
| Event                   | Would Notify                                      |
+-------------------------+---------------------------------------------------+
| order.placed            | Inventory reservation, billing system, warehouse  |
| order.payment_completed | Fulfillment service, accounting, notification     |
| order.cancelled         | Inventory release, refund service                 |
| customer.registered     | CRM, email marketing, welcome flow                |
| customer.deactivated    | Access control, subscription service              |
| product.created         | Search index, product catalog cache               |
| product.deactivated     | Search index removal, catalog cache invalidation  |
| product.activated       | Search index re-add, catalog cache update         |
+-------------------------+---------------------------------------------------+

## Summary

+------------------+-------+----------+
| Type             | Done  | Missing  |
+------------------+-------+----------+
| Domain Events    |   8   |    4     |
| Event Sourcing   |   1   |    4     |
| Integration      |   0   |    8     |
+------------------+-------+----------+

- Domain events: 4 missing cases are all rejection/failure paths
- Event sourcing: only the payment flow is append-only; all others mutate in place
- Integration events: none wired yet — LogPublisher at domain/platform/event/publisher.go is the swap point
