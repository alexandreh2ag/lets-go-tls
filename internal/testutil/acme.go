package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/platform/tester"
)

// SetupFakeAPI creates a minimal stub ACME server for testing over HTTPS.
// Unlike lego's tester.SetupFakeAPI, this registers the /nonce handler
// at the top level to avoid race conditions caused by lazy registration.
// It returns the mux, the server URL, and an *http.Client that trusts the
// test server's TLS certificate (use it with lego.Config.HTTPClient).
func SetupFakeAPI(t *testing.T) (*http.ServeMux, string, *http.Client) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("HEAD /nonce", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Replay-Nonce", "12345")
		w.Header().Set("Retry-After", "0")
	})

	mux.HandleFunc("GET /dir", func(w http.ResponseWriter, r *http.Request) {
		err := tester.WriteJSONResponse(w, acme.Directory{
			NewNonceURL:   server.URL + "/nonce",
			NewAccountURL: server.URL + "/account",
			NewOrderURL:   server.URL + "/newOrder",
			RevokeCertURL: server.URL + "/revokeCert",
			KeyChangeURL:  server.URL + "/keyChange",
			RenewalInfo:   server.URL + "/renewalInfo",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return mux, server.URL, server.Client()
}
