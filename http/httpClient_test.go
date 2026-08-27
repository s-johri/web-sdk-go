package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPostFormSendsURLEncodedPayload(t *testing.T) {
	var gotContentType, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			t.Errorf("could not parse form: %v", err)
		}
		gotBody = r.PostForm.Encode()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":1,"msg":"ok"}`))
	}))
	defer server.Close()

	response, err := PostForm(server.URL, url.Values{
		"command": {"verify_payment"},
		"var1":    {"txn001"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
	if gotBody != "command=verify_payment&var1=txn001" {
		t.Errorf("body = %q", gotBody)
	}
	if response["msg"] != "ok" {
		t.Errorf("response = %#v", response)
	}
}

func TestPostFormReturnsErrorOnNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<html><body>Bad Gateway</body></html>"))
	}))
	defer server.Close()

	response, err := PostForm(server.URL, url.Values{"command": {"verify_payment"}})
	if err == nil {
		t.Fatal("expected an error for a 502 response")
	}
	if response != nil {
		t.Errorf("expected a nil response, got %#v", response)
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("error should name the status code, got: %v", err)
	}
}

func TestPostFormHasARequestTimeout(t *testing.T) {
	if defaultClient.Timeout <= 0 {
		t.Error("the shared client must have a timeout, or a stalled PayU endpoint blocks the caller forever")
	}
}

func TestPostFormReusesOneClient(t *testing.T) {
	first := defaultClient
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":1}`))
	}))
	defer server.Close()

	if _, err := PostForm(server.URL, url.Values{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if defaultClient != first {
		t.Error("PostForm must reuse the shared client rather than build a new one per call")
	}
}
