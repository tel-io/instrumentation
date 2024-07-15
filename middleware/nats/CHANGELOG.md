== middleware/nats/v2.0.7
* update go.opentelemetry.io/otel/* dependencies according to:
    * https://github.com/open-telemetry/opentelemetry-go/blob/main/CHANGELOG.md

== middleware/nats/v2.0.2
* one simple middleware for all
* separate publisher
* ubiquitous name function
* privent races
* dashboard fixes
* simplified mw signature for both sub and pub approaches
* coverage for logs
* baggage flow
* package function WrapKindOfContext helper put to mw kind - only for none wrapped functions
* rethink log output - give logs informative
* WithDisableDefaultMiddleware option allow not use pre-build middlewares
* WithDump replace WithDumpResponse, WithDumpRequest
* Core/consumer: QueueSubscribeSyncWithChan sync function. Example shows actual usage