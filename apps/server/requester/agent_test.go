package requester

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	mockHttp "github.com/alexandreh2ag/lets-go-tls/mocks/http"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
)

func Test_agent_ID(t *testing.T) {
	want := "foo"
	instance := &agent{id: want}
	assert.Equalf(t, want, instance.ID(), "ID()")
}

func Test_createAgentProvider(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &agent{id: "foo", logger: ctx.Logger}
	want.addresses = []string{"http://127.0.0.1:80", "http://127.0.0.1:81"}
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
				Config: map[string]interface{}{"addresses": []string{"http://127.0.0.1:80", "http://127.0.0.1:81"}},
			},
			want:    want,
			wantErr: false,
		},
		{
			name: "FailDecodeCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"addresses": "http://127.0.0.1:80"},
			},
			want:        want,
			wantErr:     true,
			errContains: "'addresses': source data must be an array or slice, got string",
		},
		{
			name: "FailValidateCfg",
			cfg: config.RequesterConfig{
				Id:     "foo",
				Config: map[string]interface{}{"addresses": []string{}},
			},
			want:        want,
			wantErr:     true,
			errContains: "Key: 'ConfigAgent.Addresses' Error:Field validation for 'Addresses' failed on the 'min' tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createAgentProvider(ctx, tt.cfg)

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

func Test_agent_FetchAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	a := &agent{}
	want := []types.DomainRequest{
		{Domains: types.Domains{types.Domain("foo.com")}},
		{Domains: types.Domains{types.Domain("bar.com")}},
	}
	tests := []struct {
		name    string
		want    []types.DomainRequest
		wantErr assert.ErrorAssertionFunc

		funcMock func(clientHttp *mockHttp.MockClient)
	}{
		{
			name:    "Success",
			want:    want,
			wantErr: assert.NoError,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal(want)
				resp.SetBody(body)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
		{
			name:    "FailDoRequest",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail"))
			},
		},
		{
			name:    "FailStatusCode",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusInternalServerError)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
		{
			name:    "FailUnmarshal",
			want:    []types.DomainRequest{},
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				resp.SetBody([]byte("{]"))
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHttp := mockHttp.NewMockClient(ctrl)
			tt.funcMock(clientHttp)
			a.httpClient = clientHttp

			got, err := a.FetchAgent("127.0.0.1:80")
			if !tt.wantErr(t, err, fmt.Sprintf("FetchInstance(127.0.0.1:80)")) {
				return
			}
			requests := []types.DomainRequest{}
			for _, request := range got {
				requests = append(requests, *request)
			}
			assert.Equalf(t, tt.want, requests, "FetchInstance(127.0.0.1:80)")
		})
	}
}

func Test_agent_Fetch(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	a := &agent{addresses: []string{"127.0.0.1:80", "127.0.0.1:81"}, logger: ctx.Logger}
	want1 := []types.DomainRequest{
		{Requester: a, Domains: types.Domains{types.Domain("foo.com")}},
	}
	want2 := []types.DomainRequest{
		{Requester: a, Domains: types.Domains{types.Domain("bar.com")}},
	}
	want := append(want1, want2...)
	tests := []struct {
		name     string
		funcMock func(clientHttp *mockHttp.MockClient)
		want     []types.DomainRequest
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:    "Success",
			want:    want,
			wantErr: assert.NoError,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp1 := fasthttp.Response{}
				resp1.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal(want1)
				resp1.SetBody(body)

				resp2 := fasthttp.Response{}
				resp2.SetStatusCode(http.StatusOK)
				body, _ = json.Marshal(want2)
				resp2.SetBody(body)

				gomock.InOrder(
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp1).Return(nil),
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp2).Return(nil),
				)

			},
		},
		{
			name:    "Fail",
			want:    want1,
			wantErr: assert.Error,
			funcMock: func(clientHttp *mockHttp.MockClient) {
				resp1 := fasthttp.Response{}
				resp1.SetStatusCode(http.StatusOK)
				body, _ := json.Marshal(want1)
				resp1.SetBody(body)
				gomock.InOrder(
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp1).Return(nil),
					clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail")),
				)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHttp := mockHttp.NewMockClient(ctrl)
			tt.funcMock(clientHttp)
			a.httpClient = clientHttp
			got, err := a.Fetch()
			if !tt.wantErr(t, err, fmt.Sprintf("Fetch()")) {
				return
			}

			requests := []types.DomainRequest{}
			for _, request := range got {
				requests = append(requests, *request)
			}
			assert.ElementsMatchf(t, tt.want, requests, "Fetch()")
		})
	}
}
