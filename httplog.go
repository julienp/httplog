package httplog

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// LogRecord warps a http.ResponseWriter and records the status
type LogRecord struct {
	http.ResponseWriter
	status int
}

func (r *LogRecord) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

// WriteHeader overrides ResponseWriter.WriteHeader to keep track of the response code
func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// WithHTTPLogging adds HTTP request logging to the Handler h
func WithHTTPLogging(log *logrus.Entry) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			record := &LogRecord{
				ResponseWriter: w,
				status:         200,
			}
			h.ServeHTTP(record, r)

			level := logrus.InfoLevel
			if record.status >= 400 {
				level = logrus.WarnLevel
			}
			if record.status >= 500 {
				level = logrus.ErrorLevel
			}
			hFields := map[string]interface{}{
				"ident":       r.Host,
				"method":      r.Method,
				"referer":     r.Referer(),
				"request_id":  r.Header.Get("X-Request-Id"),
				"status_code": record.status,
				"url":         r.URL.Path,
				"useragent":   r.UserAgent(),
				"version":     fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor),
			}

			log.WithFields(logrus.Fields{"http": hFields}).Log(level)
		})
	}
}
