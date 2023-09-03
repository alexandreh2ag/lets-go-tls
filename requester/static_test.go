package requester

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_static_ID_Success(t *testing.T) {
	staticP := &static{id: "foo"}
	got := staticP.ID()

	assert.Equal(t, "foo", got)
}

func Test_static_Fetch_Success(t *testing.T) {
	staticP := &static{}
	domainRequests := []*types.DomainRequest{{Domains: types.Domains{"foo.com"}, Requester: staticP}}
	staticP.domainRequests = domainRequests
	got, err := staticP.Fetch()
	assert.Nil(t, err)
	assert.Equal(t, domainRequests, got)
}

func Test_createStaticProvider(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &static{id: "foo"}
	want.domainRequests = []*types.DomainRequest{{Domains: types.Domains{"foo.com"}, Requester: want}}
	tests := []struct {
		name        string
		cfg         config.RequesterConfig
		want        types.Requester
		wantErr     bool
		errContains string
	}{
		{
			name: "Success",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"domains": [][]string{{"foo.com"}}},
			},
			want:    want,
			wantErr: false,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"domains": "foo.com"},
			},
			want:        want,
			wantErr:     true,
			errContains: "'domains': source data must be an array or slice, got string",
		},
		{
			name: "FailValidateCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"domains": []string{}},
			},
			want:        want,
			wantErr:     true,
			errContains: "Key: 'ConfigStatic.ListDomains' Error:Field validation for 'ListDomains' failed on the 'min' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createStaticProvider(ctx, tt.cfg)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
