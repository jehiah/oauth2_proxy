package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoggingHandler_Hijack(t *testing.T) {
	// httputil.ReverseProxy hijacks the connection to proxy protocol upgrades
	// such as websockets, so responseLogger must not hide http.Hijacker from
	// the underlying ResponseWriter
	hijackErr := make(chan error, 1)
	handler := func(w http.ResponseWriter, req *http.Request) {
		conn, _, err := http.NewResponseController(w).Hijack()
		if err == nil {
			_ = conn.Close()
		}
		hijackErr <- err
	}

	h := LoggingHandler(bytes.NewBuffer(nil), http.HandlerFunc(handler), true, defaultRequestLoggingFormat)
	server := httptest.NewServer(h)
	defer server.Close()

	// the connection is hijacked and closed without a response, so a client
	// side error is expected here; only the hijack result is under test
	if resp, err := http.Get(server.URL); err == nil {
		_ = resp.Body.Close()
	}

	if actual := <-hijackErr; actual != nil {
		t.Errorf("Hijack() returned %v, expected nil", actual)
	}
}

func TestLoggingHandler_ServeHTTP(t *testing.T) {
	ts := time.Now()

	tests := []struct {
		Format,
		ExpectedLogMessage string
	}{
		{defaultRequestLoggingFormat, fmt.Sprintf("127.0.0.1 - - [%s] test-server GET - \"/foo/bar\" HTTP/1.1 \"\" 200 4 0.000\n", ts.Format("02/Jan/2006:15:04:05 -0700"))},
		{"{{.RequestMethod}}", "GET\n"},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		handler := func(w http.ResponseWriter, req *http.Request) {
			if _, err := w.Write([]byte("test")); err != nil {
				t.Fatalf("failed to write response: %v", err)
			}
		}

		h := LoggingHandler(buf, http.HandlerFunc(handler), true, test.Format)

		r, _ := http.NewRequest("GET", "/foo/bar", nil)
		r.RemoteAddr = "127.0.0.1"
		r.Host = "test-server"

		h.ServeHTTP(httptest.NewRecorder(), r)

		actual := buf.String()
		if actual != test.ExpectedLogMessage {
			t.Errorf("Log message was\n%s\ninstead of expected \n%s", actual, test.ExpectedLogMessage)
		}
	}
}
