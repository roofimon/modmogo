# modmono

HTTP API service built with [Fiber](https://gofiber.io/) and MongoDB.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24 or newer
- [Docker](https://docs.docker.com/get-docker/) (optional, for running MongoDB locally)
- [Bun](https://bun.sh/) 1.2+ (for installing and running scripts in the Angular web UI in [`web/`](web/))

## Run MongoDB

From the repository root:

```bash
docker compose up -d
```

This starts MongoDB 7 on port **27017** (see [`docker-compose.yml`](docker-compose.yml)). The app defaults to `mongodb://localhost:27017`, so no extra configuration is required for local development. Collections include `products` and `customers`.

## Run the server

```bash
go run ./cmd/server
```

The API listens on **:8080** by default. Product routes are under `/products` and customer routes under `/customers` (for example `GET http://localhost:8080/products` and `GET http://localhost:8080/customers`). See [`api/openapi.yaml`](api/openapi.yaml) for the full contract.

The process **does not** connect to MongoDB at startup. The client is created on first use (for example the first `/health/ready` check or `/products` request), so the HTTP server can start even if Mongo is still coming up. The first such call may take longer while it dials and pings the cluster.

Stop the server with `Ctrl+C` (graceful shutdown).

## Health checks

| Endpoint | Role |
|----------|------|
| `GET /health/live` | **Liveness** — the process is running and serving HTTP. Use this to decide whether to restart the container. |
| `GET /health/ready` | **Readiness** — obtains the Mongo client (connecting if this is the first use), then pings within a short timeout. Returns `503` with body `not ready` if the database is unavailable. Use this to decide whether to send application traffic to the instance (for example Kubernetes readiness probes). |

## Configuration

| Variable    | Default                         | Description        |
|------------|---------------------------------|--------------------|
| `MONGO_URI` | `mongodb://localhost:27017`    | MongoDB connection URI |
| `MONGO_DB`  | `modmono`                       | Database name      |
| `HTTP_ADDR` | `:8080`                         | Listen address     |
| `CORS_ALLOWED_ORIGINS` | *(empty)*            | Comma-separated browser origins allowed for cross-origin API calls (e.g. the Angular dev server). When empty, no CORS middleware is enabled. |

Example:

```bash
HTTP_ADDR=:3000 MONGO_DB=mydb go run ./cmd/server
```

Local SPA example (Angular on port 4200 calling the API on 8080):

```bash
CORS_ALLOWED_ORIGINS=http://localhost:4200 go run ./cmd/server
```

## Web UI (Angular)

The [`web/`](web/) app is a standalone Angular **19** SPA with a Coinbase Design System–inspired theme (SCSS tokens and Inter; not the React `@coinbase/cds-web` package). A **left sidebar** groups **Products** (catalog, inactive list, add) and **Customers** (directory, inactive list, add). It calls the APIs in [`api/openapi.yaml`](api/openapi.yaml): products use `GET /products`, `GET /products/inactive`, `GET /products/{id}`, `POST /products`, and `POST /products/{id}/deactivate`; customers use the parallel `/customers` routes.

From the repository root:

```bash
cd web
bun install
bun run start
```

By default `bun run start` (`ng serve`) runs at **http://localhost:4200**. Development uses relative API URLs and [`web/proxy.conf.json`](web/proxy.conf.json) to forward `/products`, `/customers`, and `/health` to **http://localhost:8080**, so you do **not** need CORS for local UI development as long as the API is listening on 8080.

If you point the SPA at the API with an absolute URL (e.g. `apiBaseUrl: 'http://localhost:8080'`) instead of the proxy, set `CORS_ALLOWED_ORIGINS` to the exact origin you use in the browser (including scheme and host, e.g. `http://localhost:4200` vs `http://127.0.0.1:4200`). A mismatch there often surfaces in Angular as **status 0** / “unknown” errors.

Production builds use `src/environments/environment.ts` with `apiBaseUrl` **`''`**, so requests are same-origin relative URLs (e.g. `/products`, `/customers`). Point that at your API with a reverse proxy or change the value for your deployment.

```bash
cd web
bun run build
```

Output is under `web/dist/web/`.

## Tests

### Run all tests

```bash
go test ./...
```

### Run product domain tests only

```bash
go test ./internal/product/...
```

### Run with verbose output

```bash
go test -v ./internal/product/...
```

Test coverage spans all three domains — product, customer, and order — across four layers each:

| Domain | Layer | File | What is tested |
|--------|-------|------|----------------|
| product | Pure logic | `application/id_test.go` | `parseObjectID` |
| product | Pure logic + orchestration | `application/service_test.go` | `validateCreateInput`, `buildProduct`, `Service` methods with mock repo |
| product | HTTP handlers | `adapter/http/handlers_test.go` | `parseLimit`, `createErrorToHTTP`, `idErrorToHTTP` |
| product | API (routes) | `adapter/http/api_test.go` | All routes via `fiber.App.Test()` with mock `port.UseCase` |
| customer | Pure logic | `application/id_test.go` | `parseObjectID` |
| customer | Pure logic + orchestration | `application/service_test.go` | `validateCreateInput`, `buildCustomer`, `Service` methods with mock repo |
| customer | HTTP handlers | `adapter/http/handlers_test.go` | `parseLimit`, `createErrorToHTTP`, `idErrorToHTTP` |
| customer | API (routes) | `adapter/http/api_test.go` | All routes via `fiber.App.Test()` with mock `port.UseCase` |
| order | Domain | `domain/model_test.go` | `ComputeTotal` |
| order | Domain | `domain/view_test.go` | `ToOrderView` — field mapping, enrichment, unknown SKU |
| order | Pure logic | `application/id_test.go` | `parseObjectID` |
| order | Orchestration | `application/service_test.go` | `PlaceOrder` validation, `ViewOrderDetail`, `CompletePayment`, `Deactivate` with mock repo and catalogs |
| order | HTTP handlers | `adapter/http/handlers_test.go` | `parseLimit`, `createErrorToHTTP` (5 errors), `idErrorToHTTP` (400/404/409) |
| order | API (routes) | `adapter/http/api_test.go` | All 9 routes via `fiber.App.Test()` with mock `port.UseCase` and catalogs |

## Build

```bash
go build -o modmono ./cmd/server
./modmono
```