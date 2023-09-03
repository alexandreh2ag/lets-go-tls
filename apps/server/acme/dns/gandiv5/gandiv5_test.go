package gandiv5

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_gandiV5_ID(t *testing.T) {
	id := "foo"
	p := gandiV5{id: id}
	assert.Equal(t, KeyDnsGandiV5+"-"+id, p.ID())
}

func Test_gandiV5_Type(t *testing.T) {
	p := gandiV5{}
	assert.Equal(t, acme.TypeDNS01, p.Type())
}

func Test_CreateGandiV5(t *testing.T) {
	ctx := context.TestContext(nil)
	tests := []struct {
		name        string
		cfg         map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg:  map[string]interface{}{"api_key": "api_key"},
		},
		{
			name:        "FailDecodeCfg",
			cfg:         map[string]interface{}{"api_key": []string{}},
			wantErr:     true,
			errContains: "'api_key' expected type 'string', got unconvertible type '[]string'",
		},
		{
			name:        "FailValidateCfg",
			cfg:         map[string]interface{}{"api_key": ""},
			wantErr:     true,
			errContains: "Key: 'ConfigGandiV5.APIKey' Error:Field validation for 'APIKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateGandiV5(ctx, "foo", tt.cfg)

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
