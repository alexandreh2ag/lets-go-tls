package testutil

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/stretchr/testify/assert"
)

func TestSetupFakeAPI_DirEndpoint(t *testing.T) {
	_, apiURL, httpClient := SetupFakeAPI(t)

	resp, err := httpClient.Get(apiURL + "/dir")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var dir acme.Directory
	err = json.NewDecoder(resp.Body).Decode(&dir)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, apiURL+"/nonce", dir.NewNonceURL)
	assert.Equal(t, apiURL+"/account", dir.NewAccountURL)
	assert.Equal(t, apiURL+"/newOrder", dir.NewOrderURL)
	assert.Equal(t, apiURL+"/revokeCert", dir.RevokeCertURL)
	assert.Equal(t, apiURL+"/keyChange", dir.KeyChangeURL)
	assert.Equal(t, apiURL+"/renewalInfo", dir.RenewalInfo)
}

func TestSetupFakeAPI_DirEndpointMethodNotAllowed(t *testing.T) {
	_, apiURL, httpClient := SetupFakeAPI(t)

	resp, err := httpClient.Post(apiURL+"/dir", "", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	resp.Body.Close()
}

func TestSetupFakeAPI_NonceEndpoint(t *testing.T) {
	_, apiURL, httpClient := SetupFakeAPI(t)

	req, err := http.NewRequest(http.MethodHead, apiURL+"/nonce", nil)
	assert.NoError(t, err)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "12345", resp.Header.Get("Replay-Nonce"))
	assert.Equal(t, "0", resp.Header.Get("Retry-After"))
	resp.Body.Close()
}

func TestSetupFakeAPI_NonceEndpointMethodNotAllowed(t *testing.T) {
	_, apiURL, httpClient := SetupFakeAPI(t)

	resp, err := httpClient.Get(apiURL + "/nonce")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	resp.Body.Close()
}

func TestSetupFakeAPI_NonceAvailableBeforeDir(t *testing.T) {
	_, apiURL, httpClient := SetupFakeAPI(t)

	// Nonce must be available without calling /dir first
	req, err := http.NewRequest(http.MethodHead, apiURL+"/nonce", nil)
	assert.NoError(t, err)

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "12345", resp.Header.Get("Replay-Nonce"))
	resp.Body.Close()
}

func TestSetupFakeAPI_ReturnsMuxURLAndClient(t *testing.T) {
	mux, apiURL, httpClient := SetupFakeAPI(t)

	assert.NotNil(t, mux)
	assert.NotEmpty(t, apiURL)
	assert.True(t, strings.HasPrefix(apiURL, "https://"), "URL should use HTTPS")
	assert.NotNil(t, httpClient)
}
