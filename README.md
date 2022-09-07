# Instrumentation

Repository store plugins for [tel](http://github.com/d7561985/tel) project which currently located

Take look on
opentelemetry-go-contrib: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/README.md

## Instrumentation Packages

|              Instrumentation Package              | Metrics | Traces | Logs |
|:-------------------------------------------------:|:-------:|:------:|:----:|
|   [github.com/go-chi/chi/v5](./middleware/chi)    |    ✓    |   ✓    |  ✓   |
| [github.com/labstack/echo/v4](./middleware/echo)  |    ✓    |   ✓    |  ✓   |
|   [github.com/gin-gonic/gin](./middleware/gin)    |    ✓    |   ✓    |  ✓   |
|    [google.golang.org/grpc](./middleware/grpc)    |    ✓    |   ✓    |  ✓   |
|           [net/http](./middleware/http)           |    ✓    |   ✓    |  ✓   |
| [github.com/nats-io/nats.go](./middleware/natsmw) |    ✓    |   ✓    |  ✓   |
|         [database/sql](./plugins/otelsql)         |    ✓    |   ✓    |      |
|     [github.com/jackc/pgx/v4](./plugins/pgx)      |         |        |  ✓   |
