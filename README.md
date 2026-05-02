# Hermes

High-performance gRPC microservices platform built with Go. Provides a foundation for inter-service communication with binary serialization, HTTP/2, streaming, mTLS, and observability.

## Services

| Service | Description |
|---|---|
| `user` | CRUD + server-side streaming of user events |
| `order` | CRUD, cancellation + bidirectional streaming of order status |

## Project Structure

```
Hermes/
├── proto/               # Protobuf definitions (buf toolchain)
│   ├── common/v1/       # Shared types: pagination, errors
│   ├── user/v1/         # UserService proto
│   └── order/v1/        # OrderService proto
├── gen/go/              # Generated Go code (do not edit manually)
├── services/
│   ├── user/            # User microservice
│   └── order/           # Order microservice
├── pkg/                 # Shared Go packages
├── go.work              # Go workspace
└── Makefile
```

## How It Works

### Proto-first Development

All service contracts are defined in `.proto` files under `proto/`. The source of truth is always the proto definition — Go code is generated from it, never written by hand.

```
proto/user/v1/user.proto
        ↓  make proto-gen (buf generate)
gen/go/user/v1/user.pb.go          ← message types
gen/go/user/v1/user_grpc.pb.go     ← server/client interfaces
```

Each service imports the generated code and implements the server interface. The client side gets a type-safe stub automatically — no HTTP routing, no JSON marshaling, no boilerplate.

### Go Workspace (go.work)

The repo is a monorepo with three independent Go modules:

```
pkg/            → github.com/gulmix/hermes/pkg
services/user/  → github.com/gulmix/hermes/services/user
services/order/ → github.com/gulmix/hermes/services/order
```

`go.work` links them together locally so services can import shared `pkg/` packages without publishing to a registry. In CI/production each module is built independently.

### Interceptor Pipeline

Every gRPC call (both server and client side) passes through a chain of interceptors before reaching the handler. Planned stack, in order:

```
request
  → recovery       (panic → gRPC Internal error, prevents crash)
  → auth/mTLS      (validates peer certificate)
  → logging        (structured log: method, duration, status)
  → metrics        (Prometheus counters/histograms per method)
  → tracing        (OpenTelemetry span, propagates trace-id)
  → handler
```

Interceptors are defined once in `pkg/interceptor/` and reused by every service.

### Communication Patterns

The platform covers all four gRPC patterns:

**Unary** — standard request/response, like HTTP. Used for CRUD operations (`GetUser`, `CreateOrder`, etc.).

```
client → request → server → response → client
```

**Server streaming** — server sends a stream of messages after one request. Used in `WatchUsers`: client subscribes to a status filter and receives events (created/updated/deleted) as they happen.

```
client → WatchUsersRequest → server
                              ↓ UserEvent (CREATED)
                              ↓ UserEvent (UPDATED)
                              ↓ UserEvent (DELETED)
                              ...
```

**Bidirectional streaming** — both sides send independently over one connection. Used in `StreamOrderUpdates`: client sends order IDs it wants to track, server pushes status changes back in real time.

```
client → OrderStatusRequest(order_1) →
client → OrderStatusRequest(order_2) →     server
                             ← OrderStatusResponse(order_1, CONFIRMED)
                             ← OrderStatusResponse(order_2, SHIPPED)
```

### Pagination

All list endpoints use the shared `common.v1.PageRequest` / `PageResponse` types:

```protobuf
// request
PageRequest { page: 1, page_size: 20 }

// response
PageResponse { total: 143, page: 1, page_size: 20, has_next: true }
```

### Error Handling

The `common.v1.AppError` type carries a typed error code, a human-readable message, and an optional field name for validation errors. This is returned as gRPC status detail metadata alongside the standard gRPC status code.

```
ErrorCode: NOT_FOUND / ALREADY_EXISTS / INVALID_ARGUMENT / PERMISSION_DENIED / INTERNAL
```

### mTLS

Service-to-service calls are authenticated via mutual TLS — both sides present a certificate. This means:
- No service can be impersonated — every connection is verified by a shared CA
- Traffic is encrypted end-to-end
- No separate auth token needed between internal services

### Observability

Three pillars, all correlated by `trace-id`:

| Signal | Tool | What you see |
|---|---|---|
| Metrics | Prometheus + Grafana | RPC latency (p50/p99), error rate, in-flight calls per method |
| Traces | OpenTelemetry → Jaeger | Full call chain across services, per-span timing |
| Logs | zap (structured JSON) | Every request: method, duration, status code, trace-id |

The `trace-id` is injected by the tracing interceptor and attached to every log line, so you can jump from a Grafana alert → Jaeger trace → exact log lines.

---

## gRPC API

### UserService

| Method | Type | Description |
|---|---|---|
| `GetUser` | Unary | Get user by ID |
| `ListUsers` | Unary | Paginated list with status filter |
| `CreateUser` | Unary | Create new user |
| `UpdateUser` | Unary | Update display name or status |
| `DeleteUser` | Unary | Delete user by ID |
| `WatchUsers` | Server streaming | Stream user events (created/updated/deleted) |

### OrderService

| Method | Type | Description |
|---|---|---|
| `GetOrder` | Unary | Get order by ID |
| `ListOrders` | Unary | Paginated list by user with status filter |
| `CreateOrder` | Unary | Create order with items |
| `CancelOrder` | Unary | Cancel order with reason |
| `StreamOrderUpdates` | Bidirectional streaming | Real-time order status updates |

---

## Prerequisites

- Go 1.26+
- [buf](https://buf.build/docs/installation) v1.47+
- `protoc-gen-go` and `protoc-gen-go-grpc`

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Makefile Commands

```bash
make proto-gen       # Generate Go code from proto files
make proto-lint      # Lint proto files with buf
make proto-format    # Format proto files in place
make proto-breaking  # Check for breaking changes against main
make proto-all       # lint + generate
make deps-update     # Update buf dependencies
```

## CI

GitHub Actions runs on every push/PR that touches `proto/`:

- **Lint** — `buf lint`
- **Format check** — `buf format --diff --exit-code`
- **Breaking change detection** — compared against the base branch (PRs only)
- **Generated code verification** — ensures `gen/` is up to date with the proto definitions

## Key Dependencies

| Package | Purpose |
|---|---|
| `google.golang.org/grpc` | gRPC runtime |
| `google.golang.org/protobuf` | Protobuf serialization |
| `buf` | Proto toolchain |
| `go.opentelemetry.io/otel` | Distributed tracing |
| `github.com/prometheus/client_golang` | Metrics |
| `go.uber.org/zap` | Structured logging |
