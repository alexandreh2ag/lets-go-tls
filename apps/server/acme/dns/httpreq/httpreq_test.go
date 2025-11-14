package httpreq

import (
	"testing"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/stretchr/testify/assert"
)

func Test_httpReq_ID(t *testing.T) {
	id := "foo"
	p := httpReq{id: id}
	assert.Equal(t, KeyDnsHttpReq+"-"+id, p.ID())
}

func Test_httpReq_Type(t *testing.T) {
	p := httpReq{}
	assert.Equal(t, acme.TypeDNS01, p.Type())
}

func Test_CreatehttpReq(t *testing.T) {
	ctx := context.TestContext(nil)
	tests := []struct {
		name        string
		cfg         map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg:  map[string]interface{}{"endpoint": "http://127.0.0.1/", "mode": ""},
		},
		{
			name: "SuccessWithDuration",
			cfg:  map[string]interface{}{"endpoint": "http://127.0.0.1/", "mode": "", "http_timeout": "10s"},
		},
		{
			name:        "FailDecodeCfg",
			cfg:         map[string]interface{}{"endpoint": []string{}},
			wantErr:     true,
			errContains: "'endpoint' expected type 'string', got unconvertible type '[]string'",
		},
		{
			name:        "FailParseEndpoint",
			cfg:         map[string]interface{}{"endpoint": "http://wrong:wrong"},
			wantErr:     true,
			errContains: "invalid port \":wrong\" after host",
		},
		{
			name:        "FailValidateCfg",
			cfg:         map[string]interface{}{"endpoint": ""},
			wantErr:     true,
			errContains: "Key: 'ConfigHttpReq.Endpoint' Error:Field validation for 'Endpoint' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateHttpReq(ctx, "foo", tt.cfg)

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
