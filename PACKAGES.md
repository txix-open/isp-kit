# Package Tree

```
isp-kit                        Microservice framework 
├── app                        Application lifecycle management
├── bgjobx                     Background job client with worker management
│   ├── handler                Background job handler middlewares
│   └── migration              SQL migration files for bgjobs
├── bootstrap                  App bootstrap: config, logging, health, observability
├── cluster                    Distributed session management (isp-config-service integration)
├── codec                      Payload encoding/decoding with Zstd compression
├── common_endpoints           Common endpoints (Swagger, etc.)
├── config                     Multi-source config with env overrides
├── db                         PostgreSQL client wrapper (sqlx + pgx)
│   ├── jsonb                  JSONB PostgreSQL type for pgx
│   └── query                  Fluent SQL builder (Squirrel)
├── dbrx                       Dynamic DB client with hot-reload
├── dbx                        PostgreSQL client with pooling & tracing
│   └── migration              DB migration runner (goose)
├── errors                     Error creation, wrapping, inspection helpers
├── grmqx                      RabbitMQ client with topology, metrics, tracing
│   ├── batch_handler          Batch message processing handler
│   └── handler                RabbitMQ consumer middlewares
├── grpc                       gRPC server with middleware & auth
│   ├── apierrors              gRPC error codes & status conversion
│   ├── client                 gRPC client with LB & middleware
│   │   └── request            gRPC request builder
│   ├── endpoint               gRPC caller via reflection (wrapper)
│   │   └── grpclog            gRPC request/response logging
│   └── isp                    Generated protobuf definitions
├── healthcheck                Component health check registry
├── http                       HTTP server with handlers & middleware
│   ├── apierrors              HTTP structured error handling
│   ├── endpoint               HTTP handler caller via reflection (wrapper)
│   │   ├── buffer             ResponseWriter body buffering
│   │   └── httplog            HTTP request/response logging
│   ├── httpcli                HTTP client
│   ├── httpclix               HTTP client with default middlewares
│   ├── router                 HTTP routing (httprouter)
│   └── soap                   SOAP action multiplexer
│       └── client             SOAP client for web services
├── infra/pprof                pprof profiling endpoint
├── json                       High-performance JSON (jsoniter)
├── kafkax                     Kafka client with auth, TLS, config
│   ├── consumer               Kafka consumer
│   ├── handler                Kafka consumer middlewares
│   └── publisher              Kafka message publisher
├── lb                         Round-robin load balancer
├── log                        Logger (zap adapter)
│   ├── file                   Log rotation (lumberjack)
│   └── logutil                Log level utilities
├── metrics                    Prometheus metrics utilities
│   ├── app_metrics            App-level log sampling metrics
│   ├── bgjob_metrics          Background job latency/count metrics
│   ├── db_metrics             DB connection pool metrics
│   ├── grpc_metrics           gRPC request latency metrics
│   ├── http_metrics           HTTP request latency metrics
│   ├── kafka_metrics          Kafka consumer/publisher metrics
│   ├── pgx_metrics            pgxpool connection pool metrics
│   ├── rabbitmq_metrics       RabbitMQ consumer/publisher metrics
│   └── sql_metrics            SQL query duration tracing
├── observability              Observability utilities
│   ├── sentry                 Sentry error tracking
│   └── tracing                OpenTelemetry distributed tracing
│       ├── grpc               gRPC metadata carrier for OTel
│       ├── grpc/client_tracing    gRPC client tracing middleware
│       ├── grpc/server_tracing    gRPC server tracing middleware
│       ├── http                   HTTP tracing subpackages
│       ├── http/client_tracing    HTTP client tracing middleware
│       ├── http/semconvutil     OTel HTTP semantic conventions
│       ├── http/server_tracing  HTTP server tracing middleware
│       ├── rabbitmq             RabbitMQ AMQP OTel carrier
│       ├── rabbitmq/consumer_tracing  RMQ consumer tracing middleware
│       ├── rabbitmq/publisher_tracing RMQ publisher tracing middleware
│       └── sql_tracing        pgx SQL query tracing
├── panic_recovery             Panic recovery helper
├── rc                         Remote config with hot-reload & validation
│   └── schema                 JSON schema generation for config
├── requestid                  Request ID context management
├── retry                      Exponential backoff retry
├── shutdown                   Graceful shutdown signal handling
├── stompx                     STOMP protocol wrapper
│   ├── consumer               STOMP message consumer
│   ├── handler                STOMP consumer middlewares
│   └── publisher              STOMP message publisher
├── test                       Test helper utilities
│   ├── dbt                    DB test helpers with schema cleanup
│   ├── fake                   Fake data generation
│   ├── grmqt                  RabbitMQ test helpers
│   ├── grpct                  gRPC mock test helpers
│   ├── httpt                  HTTP test server helpers
│   ├── kafkat                 Kafka test helpers
│   ├── rct                    Remote config validation helpers
│   └── stompt                 STOMP test helpers
├── validator                  Struct validation (go-playground/validator)
└── worker                     Periodic task executor with concurrency control
```
