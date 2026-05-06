# gRPC Microservices — План проекта

## Цель

Высокопроизводительный gRPC-сервис для взаимодействия между микросервисами:
бинарная сериализация, HTTP/2, двунаправленный стриминг, TLS, мониторинг.

---

## Фазы

### Фаза 4 — Паттерны взаимодействия

- [ ] Унарный RPC (Unary) — стандартный запрос/ответ
- [ ] Server-side streaming — потоки событий от сервера
- [ ] Client-side streaming — загрузка батчей
- [ ] Bidirectional streaming — реалтайм-взаимодействие
- [ ] Deadline propagation по всей цепочке вызовов

### Фаза 5 — Производительность

- [ ] Настроить keepalive параметры (`grpc.KeepaliveParams`)
- [ ] Сжатие сообщений (`gzip` / `snappy`)
- [ ] Connection pooling на клиенте
- [ ] Бенчмарки через `ghz` и встроенный `testing.B`
- [ ] Профилировка через `pprof`-endpoint

### Фаза 6 — Наблюдаемость

- [ ] Метрики Prometheus: latency, error rate, in-flight RPC
- [ ] Трассировка OpenTelemetry → Jaeger/Tempo
- [ ] Структурированные логи с trace-id корреляцией
- [ ] Grafana-дашборд для gRPC-метрик

### Фаза 7 — Деплой

- [ ] Dockerfile для каждого сервиса (multi-stage build)
- [ ] docker-compose для локальной разработки
- [ ] Kubernetes-манифесты:
  - Deployment, Service (ClusterIP)
  - ConfigMap / Secret для TLS-сертификатов
  - HorizontalPodAutoscaler
- [ ] Envoy/nginx как gRPC-прокси (опционально)

---

## Структура директорий

```
grpc-microservices/
├── proto/
│   ├── buf.yaml
│   ├── buf.gen.yaml
│   ├── common/v1/
│   ├── user/v1/
│   └── order/v1/
├── gen/                    # сгенерированный код (gitignore или нет)
│   └── go/
├── services/
│   ├── user/               # User-сервис
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── repository/
│   │   │   └── service/
│   │   └── Dockerfile
│   └── order/              # Order-сервис
│       ├── cmd/server/
│       ├── internal/
│       └── Dockerfile
├── pkg/
│   ├── grpcserver/         # общая обёртка сервера
│   ├── grpcclient/         # общая обёртка клиента
│   ├── interceptor/        # переиспользуемые interceptors
│   ├── telemetry/          # OTEL setup
│   └── tlsconfig/          # TLS helpers
├── deploy/
│   ├── docker-compose.yaml
│   └── k8s/
├── tests/
│   └── bench/              # gRPC бенчмарки
├── Makefile
└── go.work                 # Go workspace для монорепо
```

---

## Ключевые зависимости

| Пакет | Назначение |
|---|---|
| `google.golang.org/grpc` | gRPC runtime |
| `google.golang.org/protobuf` | Protobuf сериализация |
| `github.com/bufbuild/buf` | Proto toolchain |
| `go.opentelemetry.io/otel` | Трассировка |
| `github.com/prometheus/client_golang` | Метрики |
| `go.uber.org/zap` | Логирование |
| `github.com/grpc-ecosystem/go-grpc-middleware` | Interceptor набор |
| `google.golang.org/grpc/health` | Health-check |

---

## Makefile — базовые команды

```makefile
proto-gen:     # buf generate
proto-lint:    # buf lint
build:         # go build ./...
test:          # go test ./...
bench:         # go test -bench=. ./tests/bench/...
docker-up:     # docker-compose up
```

---

## Критерии готовности MVP

- Два сервиса общаются через gRPC с mTLS
- Latency p99 < 5ms при 1000 RPC/s на локальной машине
- Метрики доступны в Prometheus, трейсы — в Jaeger
- Health-check проходит в Kubernetes readiness probe
