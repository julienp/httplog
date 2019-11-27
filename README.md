# httplog

Middleware for http.Handler that logs each request to logrus.

Most of the fields match standard HTTP fields for Datadog.

Example usage with gorilla/mux:

```go
log := logrus.StandardLogger()
router := mux.NewRouter()
router.Use(WithHTTPLogging(log.WithField("service": "my-http-service")))
```