# gaphttp

handler and client middleware, and other tools

### CLIENT ROUND TRIPPER

```go
    httpClient := &http.Client{
        Transport: NewMiddlewareRoundTrip{http.DefaultTransport},
    }
```

### HTTP MIDDLEWARE

```go
        router := chi.NewRouter()
	router.Use(NewMiddleware())

	router.Post("/do", func(writer http.ResponseWriter, request *http.Request) {})
```