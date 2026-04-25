# Business Observability via Event Storming

## Core Idea

Event Storming is the discovery phase of business observability. The events named during storming (`OrderPlaced`, `OrderPaymentCompleted`, `CustomerRegistered`) become the natural units of measurement — business stakeholders already think in these terms, so metrics derived from them are immediately legible without engineering translation.

## Business Questions → Events

| Business Question | Event | Signal |
|---|---|---|
| How many orders placed per day? | `order.placed` | count |
| What is the conversion rate to paid? | `order.payment_completed` / `order.placed` | ratio |
| How many customers are we losing? | `customer.deactivated` | rate over time |
| Are products being retired faster than created? | `product.deactivated` vs `product.created` | delta |
| Average order value trend | `order.placed` → `total` field | histogram |

## Current State

Events are emitted via `LogPublisher` (structured `slog` output). To turn these into business dashboards, route events to a metrics sink:

- **Prometheus**: increment a counter on each event type, expose `/metrics`
- **Datadog / Loki**: query log stream by `event.type`
- **Analytics stream**: push to Kafka/NATS topic per event type

The `event.Publisher` interface in `internal/platform/event/publisher.go` is the swap point — implement a `PrometheusPublisher` or `StreamPublisher` without touching any domain code.

## What Event Storming Also Reveals (Gaps)

Missing events = missing observability:

- No `OrderValidationFailed` — lost intent (customer tried to order but couldn't) is invisible
- No `CustomerActivated` — customer reactivation cannot be measured (also not implemented)
- `OrderStatus = ""` for pending — fragile; a missing status looks the same as intentional pending

These hot spots should become explicit events before adding dashboards, otherwise the metrics will silently misrepresent reality.
