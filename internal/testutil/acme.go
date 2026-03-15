package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/platform/tester"
)

// SetupFakeAPI creates a minimal stub ACME server for testing.
// Unlike lego's tester.SetupFakeAPI, this registers the /nonce handler
// at the top level to avoid race conditions caused by lazy registration.
func SetupFakeAPI(t *testing.T) (*http.ServeMux, string) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/nonce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Replay-Nonce", "12345")
		w.Header().Set("Retry-After", "0")
	})

	mux.HandleFunc("/dir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

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

	return mux, server.URL
}
