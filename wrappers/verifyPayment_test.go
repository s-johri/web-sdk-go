package wrappers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/payu-india/web-sdk-go/utils"
)

func TestVerifyPaymentPostsTheExpectedPayload(t *testing.T) {
	var got map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("could not parse form: %v", err)
		}
		got = map[string]string{
			"key":     r.PostFormValue("key"),
			"command": r.PostFormValue("command"),
			"var1":    r.PostFormValue("var1"),
			"hash":    r.PostFormValue("hash"),
		}
		w.Write([]byte(`{"status":1,"transaction_details":{}}`))
	}))
	defer server.Close()

	response, err := VerifyPayment(testCreds(), server.URL, "txn001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response["status"] != float64(1) {
		t.Errorf("response = %#v", response)
	}

	wantHash := utils.ApiHasher(testCreds(), utils.ApiStruct{Command: "verify_payment", Var1: "txn001"})
	want := map[string]string{
		"key":     "testkey",
		"command": "verify_payment",
		"var1":    "txn001",
		"hash":    wantHash,
	}
	for field, wantValue := range want {
		if got[field] != wantValue {
			t.Errorf("%s = %q, want %q", field, got[field], wantValue)
		}
	}
}

func TestVerifyPaymentPropagatesTransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if _, err := VerifyPayment(testCreds(), server.URL, "txn001"); err == nil {
		t.Error("expected an error for a 500 response")
	}
}
