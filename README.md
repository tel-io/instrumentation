# Instrumentation

Repository store plugins for [tel](http://github.com/tel-io/tel) project which currently located

Take look on
opentelemetry-go-contrib: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/README.md

## Instrumentation Packages

|               Instrumentation Package                | Metrics | Traces | Logs |
|:----------------------------------------------------:|:-------:|:------:|:----:|
|     [github.com/go-chi/chi/v5](./middleware/chi)     |    ✓    |   ✓    |  ✓   |
|   [github.com/labstack/echo/v4](./middleware/echo)   |    ✓    |   ✓    |  ✓   |
|     [github.com/gin-gonic/gin](./middleware/gin)     |    ✓    |   ✓    |  ✓   |
|     [google.golang.org/grpc](./middleware/grpc)      |    ✓    |   ✓    |  ✓   |
|            [net/http](./middleware/http)             |    ✓    |   ✓    |  ✓   |
|   [github.com/nats-io/nats.go](./middleware/nats)    |    ✓    |   ✓    |  ✓   |
|          [database/sql](./plugins/otelsql)           |    ✓    |   ✓    |      |
|       [github.com/jackc/pgx/v4](./plugins/pgx)       |         |        |  ✓   |
| [go.mongodb.org/mongo-driver/mongo](./plugins/mingo) |         |   ✓    |      |

## Grafana Dashboards

In addition, we provide specific [grafana-dashboards](./grafana-dashboards)

pinup.product.ua.api.marketing.slider.web.requested
pinup.marketing.ua.answerito.slider.responed.xxxxxxxx1


apis:
pinup.product.ua.api.marketing.slider.web.requested - product
pinup.marketing.ua.answerito.product.slider.web.responsed - marketing

product ->
pub:
    subject: pinup.product.ua.api.marketing.slider.web.requested{"id":"UID,..."}
    reply: pinup.marketing.ua.answerito.product.slider.web.responsed.<base64>

marketing ->
    sub: pinup.product.ua.api.marketing.slider.web.requested
    reply/pub: pinup.marketing.ua.answerito.slider.web.responsed.*{"quize_id":"..."}