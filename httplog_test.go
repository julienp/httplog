package httplog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestLogFields(t *testing.T) {
	log, hook := test.NewNullLogger()
	router := mux.NewRouter()
	router.Use(WithHTTPLogging(log.WithField("service", "my-http-service")))
	router.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/path", nil)
	req.Header.Add("User-Agent", "golang-test")
	req.Header.Add("Referer", "http://example.com/referrer")
	req.Header.Add("X-Request-ID", "CAFE1234")
	router.ServeHTTP(w, req)

	expected := logrus.Fields{
		"http": map[string]interface{}{
			"ident":       "example.com",
			"method":      "GET",
			"referer":     "http://example.com/referrer",
			"request_id":  "CAFE1234",
			"status_code": 200,
			"url":         "/path",
			"useragent":   "golang-test",
			"version":     "1.1",
		},
		"service": "my-http-service",
	}
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, expected, hook.LastEntry().Data)
}

func TestLogLevel(t *testing.T) {
	log, hook := test.NewNullLogger()
	router := mux.NewRouter()
	router.Use(WithHTTPLogging(log.WithField("service", "my-http-service")))
	router.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	router.HandleFunc("/warn", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	router.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
	})
	router.HandleFunc("/default", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello Word!")
	})

	tests := []struct {
		path  string
		level logrus.Level
	}{
		{"/ok", logrus.InfoLevel},
		{"/warn", logrus.WarnLevel},
		{"/error", logrus.ErrorLevel},
		{"/default", logrus.InfoLevel},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://example.com"+test.path, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 1, len(hook.Entries))
		assert.Equal(t, test.level, hook.LastEntry().Level)
		hook.Reset()
	}
}
