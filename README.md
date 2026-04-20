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

This starts MongoDB 7 on port **27017** (see [`docker-compose.yml`](docker-compose.yml)). The app defaults to `mongodb://localhost:27017`, so no extra configuration is required for local development.

## Run the server

```bash
go run ./cmd/server
```

The API listens on **:8080** by default. Product routes are under `/products` (for example `GET http://localhost:8080/products`).

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

The [`web/`](web/) app is a standalone Angular **19** SPA with a Coinbase Design System–inspired theme (SCSS tokens and Inter; not the React `@coinbase/cds-web` package). It talks to the same product API described in [`api/openapi.yaml`](api/openapi.yaml): catalog uses `GET /products` (active only); inactive inventory uses `GET /products/inactive`; detail uses `GET /products/{id}`; create uses `POST /products`; soft-deactivate uses `POST /products/{id}/deactivate`. The **Inactive** nav item opens the inactive list page.

From the repository root:

```bash
cd web
bun install
bun run start
```

By default `bun run start` (`ng serve`) runs at **http://localhost:4200**. Development uses relative API URLs and [`web/proxy.conf.json`](web/proxy.conf.json) to forward `/products` and `/health` to **http://localhost:8080**, so you do **not** need CORS for local UI development as long as the API is listening on 8080.

If you point the SPA at the API with an absolute URL (e.g. `apiBaseUrl: 'http://localhost:8080'`) instead of the proxy, set `CORS_ALLOWED_ORIGINS` to the exact origin you use in the browser (including scheme and host, e.g. `http://localhost:4200` vs `http://127.0.0.1:4200`). A mismatch there often surfaces in Angular as **status 0** / “unknown” errors.

Production builds use `src/environments/environment.ts` with `apiBaseUrl` **`''`**, so requests are same-origin relative URLs (e.g. `/products`). Point that at your API with a reverse proxy or change the value for your deployment.

```bash
cd web
bun run build
```

Output is under `web/dist/web/`.

## Build

```bash
go build -o modmono ./cmd/server
./modmono
```
