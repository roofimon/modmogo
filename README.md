# modmono

HTTP API service built with [Fiber](https://gofiber.io/) and MongoDB.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24 or newer
- [Docker](https://docs.docker.com/get-docker/) (optional, for running MongoDB locally)

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

Stop the server with `Ctrl+C` (graceful shutdown).

## Configuration

| Variable    | Default                         | Description        |
|------------|---------------------------------|--------------------|
| `MONGO_URI` | `mongodb://localhost:27017`    | MongoDB connection URI |
| `MONGO_DB`  | `modmono`                       | Database name      |
| `HTTP_ADDR` | `:8080`                         | Listen address     |

Example:

```bash
HTTP_ADDR=:3000 MONGO_DB=mydb go run ./cmd/server
```

## Build

```bash
go build -o modmono ./cmd/server
./modmono
```
