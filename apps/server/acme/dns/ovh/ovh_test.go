package ovh

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ovh_ID(t *testing.T) {
	id := "foo"
	p := ovhChallenge{id: id}
	assert.Equal(t, KeyDnsOVH+"-"+id, p.ID())
}

func Test_ovh_Type(t *testing.T) {
	p := ovhChallenge{}
	assert.Equal(t, acme.TypeDNS01, p.Type())
}

func Test_CreateOVH(t *testing.T) {
	ctx := context.TestContext(nil)
	tests := []struct {
		name        string
		cfg         map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg: map[string]interface{}{
				"application_key":    "application_key",
				"application_secret": "application_secret",
				"consumer_key":       "consumer_key",
			},
		},
		{
			name: "SuccessWithAccessToken",
			cfg: map[string]interface{}{
				"access_token": "access_token",
			},
		},
		{
			name: "SuccessWithOAuth",
			cfg: map[string]interface{}{
				"client_id":     "client_id",
				"client_secret": "client_secret",
			},
		},
		{
			name: "FailDecodeCfg",
			cfg: map[string]interface{}{
				"application_key": []string{},
			},
			wantErr:     true,
			errContains: "'application_key' expected type 'string', got unconvertible type '[]string'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateOvh(ctx, "foo", tt.cfg)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
