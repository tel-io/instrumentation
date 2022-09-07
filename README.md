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

## Modules, Plugins and ect

Some plugins require external packages, we don'ty like unnecessary increasing dependencies.
Thus offer sub-modules which should be added separately

### Middlewares

#### http





==== grpc

[source,bash]
----
go get -v github.com/d7561985/tel/middleware/grpc/v2@latest
----

server:

[source,go]
----
import(
mw "github.com/d7561985/tel/middleware/grpc/v2"
)
func main(){
server := grpc.NewServer(
grpc.UnaryInterceptor(mw.UnaryServerInterceptorAll(mw.WithTel(&tele))),
grpc.StreamInterceptor(mw.StreamServerInterceptor()),
)
}

----

client:

[source,go]
----
import(
mw "github.com/d7561985/tel/middleware/grpc/v2"
)
func main(){
conn, err := grpc.Dial(hostPort, grpc.WithTransportCredentials(insecure.NewCredentials()),
grpc.WithUnaryInterceptor(mw.UnaryClientInterceptorAll(mw.WithTel(&tele))),
grpc.WithStreamInterceptor(mw.StreamClientInterceptor()),
)
}

----

==== NATS

[source,bash]
----
go get -v github.com/d7561985/tel/middleware/natsmw/v2@latest
----

==== chi

[source,bash]
----
go get -v github.com/d7561985/tel/middleware/chi/v2@latest
----

==== echo

[source,bash]
----
go get -v github.com/d7561985/tel/middleware/echo/v2@latest
----

==== Propagators

https://opentelemetry.io/docs/reference/specification/context/api-propagators/[specification]

In few words: this is a way how trace propagate between services

.github.com/d7561985/tel/v2/propagators/natsprop
Just helper which uses any TextMapPropagator (by default globally declared or via WithPropagators option).
Suitable propagate traces (`propagation.TraceContext`) or baggage(`propagation.Baggage`).

=== Plugins

==== Logging

github.com/d7561985/tel/plugins/pgx/v2

[source,bash]
----
go get -v github.com/d7561985/tel/plugins/pgx/v2@latest
----

==== SQL
[source,bash]
----
go get -v github.com/d7561985/tel/plugins/otelsql/v2@latest
----
For documentation please visit README.md file on plugin location
