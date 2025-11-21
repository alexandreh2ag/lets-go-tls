package manager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	mockPrometheus "github.com/alexandreh2ag/lets-go-tls/mocks/prometheus"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	mockTypesStorageState "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/state"
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	typesAcme "github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
	"github.com/jonboulle/clockwork"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const certPemResponseMock = `-----BEGIN CERTIFICATE-----
MIIDEDCCAfigAwIBAgIHPhckqW5fPDANBgkqhkiG9w0BAQsFADAoMSYwJAYDVQQD
Ex1QZWJibGUgSW50ZXJtZWRpYXRlIENBIDM5NWU2MTAeFw0xODExMDcxNzQ2NTZa
Fw0yMzExMDcxNzQ2NTZaMBMxETAPBgNVBAMTCGFjbWUud3RmMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwtLNKvZXD20XPUQCWYSK9rUSKxD9Eb0c9fag
bxOxOkLRTgL8LH6yln+bxc3MrHDou4PpDUdeo2CyOQu3CKsTS5mrH3NXYHu0H7p5
y3riOJTHnfkGKLT9LciGz7GkXd62nvNP57bOf5Sk4P2M+Qbxd0hPTSfu52740LSy
144cnxe2P1aDYehrEp6nYCESuyD/CtUHTo0qwJmzIy163Sp3rSs15BuCPyhySnE3
BJ8Ggv+qC6D5I1932DfSqyQJ79iq/HRm0Fn84am3KwvRlUfWxabmsUGARXoqCgnE
zcbJVOZKewv0zlQJpfac+b+Imj6Lvt1TGjIz2mVyefYgLx8gwwIDAQABo1QwUjAO
BgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMAwG
A1UdEwEB/wQCMAAwEwYDVR0RBAwwCoIIYWNtZS53dGYwDQYJKoZIhvcNAQELBQAD
ggEBABB/0iYhmfPSQot5RaeeovQnsqYjI5ryQK2cwzW6qcTJfv8N6+p6XkqF1+W4
jXZjrQP8MvgO9KNWlvx12vhINE6wubk88L+2piAi5uS2QejmZbXpyYB9s+oPqlk9
IDvfdlVYOqvYAhSx7ggGi+j73mjZVtjAavP6dKuu475ZCeq+NIC15RpbbikWKtYE
HBJ7BW8XQKx67iHGx8ygHTDLbREL80Bck3oUm7wIYGMoNijD6RBl25p4gYl9dzOd
TqGl5hW/1P5hMbgEzHbr4O3BfWqU2g7tV36TASy3jbC3ONFRNNYrpEZ1AL3+cUri
OPPkKtAKAbQkKbUIfsHpBZjKZMU=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDDDCCAfSgAwIBAgIIOV5hkYJx0JwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVUGViYmxlIFJvb3QgQ0EgNTBmZmJkMB4XDTE4MTEwNzE3NDY0N1oXDTQ4MTEw
NzE3NDY0N1owKDEmMCQGA1UEAxMdUGViYmxlIEludGVybWVkaWF0ZSBDQSAzOTVl
NjEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCacwXN4LmyRTgYS8TT
SZYgz758npHiPTBDKgeN5WVmkkwW0TuN4W2zXhEmcM82uxOEjWS2drvK0+iJKneh
0fQR8ZF35dIYFe8WXTg3kEmqcizSgh4LxlOntsXvatfX/6GU/ADo3xAFoBKCijen
SRBIY65yq5m00cWx3RMIcQq1B0X8nJS0O1P7MYE/Vvidz5St/36RXVu1oWLeS5Fx
HAezW0lqxEUzvC+uLTFWC6f/CilzmI7SsPAkZBk7dO5Qs0d7m/zWF588vlGS+0pt
D1on+lU85Ma2zuAd0qmB6LY66N8pEKKtMk93wF/o4Z5i58ahbwNvTKAzz4JSRWSu
mB9LAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcD
AQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEA
upU0DjzvIvoCOYKbq1RRN7rPdYad39mfjxgkeV0iOF5JoIdO6y1C7XAm9lT69Wjm
iUPvnCTMDYft40N2SvmXuuMaPOm4zjNwn4K33jw5XBnpwxC7By/Y0oV+Sl10fBsd
QqXC6H7LcSGkv+4eJbgY33P4uH5ZAy+2TkHUuZDkpufkAshzBust7nDAjfv3AIuQ
wlPoyZfI11eqyiOqRzOq+B5dIBr1JzKnEzSL6n0JLNQiPO7iN03rud/wYD3gbmcv
rzFL1KZfz+HZdnFwFW2T2gVW8L3ii1l9AJDuKzlvjUH3p6bgihVq02sjT8mx+GM2
7R4IbHGnj0BJA2vMYC4hSw==
-----END CERTIFICATE-----
`

func TestNewManager(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	want := &CertifierManager{stateStorage: stateStorage, clock: clockwork.NewRealClock()}
	got := NewManager(stateStorage)
	assert.Equal(t, want, got)
}

func TestCreateManager_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.State.Type = "fs"
	ctx.Config.State.Config = map[string]interface{}{"path": "/app"}
	got, err := CreateManager(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCertifierManager_releaseLock_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	cm := &CertifierManager{}
	err := cm.releaseLock(ctx)
	assert.NoError(t, err)
}

func TestCertifierManager_obtainLock_WithMemoryCache(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	cm := &CertifierManager{
		ephemeralID: "test",
	}
	got, err := cm.obtainLock(ctx)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestCertifierManager_obtainLock(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cacheManager := mockTypes.NewMockCache[string](ctrl)
	ctx.Cache = cacheManager
	tests := []struct {
		name     string
		mockFunc func(cache *mockTypes.MockCache[string])
		want     bool
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", &store.NotFound{}),
					cacheManager.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("test", nil),
				)
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "SuccessAlreadyLock",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("other", nil),
				)
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "SuccessLockConcurrency",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", &store.NotFound{}),
					cacheManager.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("other", nil),
				)
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "FailedGetError",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("fail")),
				)
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "FailedSetError",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", &store.NotFound{}),
					cacheManager.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail")),
				)
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "FailedSecondGetError",
			mockFunc: func(cache *mockTypes.MockCache[string]) {
				gomock.InOrder(
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", &store.NotFound{}),
					cacheManager.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
					cacheManager.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("fail")),
				)
			},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CertifierManager{
				ephemeralID: "test",
			}
			tt.mockFunc(cacheManager)
			got, err := cm.obtainLock(ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("obtainLock(%v)", ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "obtainLock(%v)", ctx)
		})
	}
}

func TestCertifierManager_FetchRequests_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mockTypes.NewMockRequester(ctrl)
	request1 := &types.DomainRequest{Domains: types.Domains{"example.com"}, Requester: r}
	request2 := &types.DomainRequest{Domains: types.Domains{"*.example.com", "example.com"}, Requester: r}
	request3 := &types.DomainRequest{Domains: types.Domains{"*.bar.com", "bar.com"}, Requester: r}
	gomock.InOrder(
		r.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{request1, request2}, nil),
		r.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{request3}, nil),
	)
	ctx.Requesters = map[string]types.Requester{"test": r, "test2": r}
	mgr := &CertifierManager{}
	want := []*types.DomainRequest{request3, request2, request1}
	got, err := mgr.FetchRequests(ctx)
	assert.Len(t, err, 0)
	assert.Equal(t, want, got)
}

func TestCertifierManager_MatchingRequests(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mockTypes.NewMockRequester(ctrl)
	ctx.Requesters = map[string]types.Requester{"test": r}

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cert1 := &types.Certificate{
		Identifier: "example.com-0",
		Main:       "example.com",
		Domains:    types.Domains{types.Domain("example.com")},
	}

	tests := []struct {
		name            string
		state           *types.State
		domainsRequests []*types.DomainRequest
		wantFunc        func() types.Certificates
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:  "SuccessWithExistingCert",
			state: &types.State{Certificates: types.Certificates{cert1}},
			wantFunc: func() types.Certificates {
				cert := &types.Certificate{
					Identifier: "example.com-0",
					Main:       "example.com",
					Domains:    types.Domains{types.Domain("example.com")},
				}
				return types.Certificates{cert}
			},
			domainsRequests: []*types.DomainRequest{{Requester: r, Domains: types.Domains{types.Domain("example.com")}}},
			wantErr:         assert.NoError,
		},
		{
			name:  "SuccessNewCert",
			state: &types.State{Certificates: types.Certificates{cert1}},
			wantFunc: func() types.Certificates {
				cert1 := &types.Certificate{
					Identifier: "example.com-0",
					Main:       "example.com",
					Domains:    types.Domains{types.Domain("example.com")},
				}
				cert2 := &types.Certificate{
					Identifier: "example2.com-0",
					Main:       "example2.com",
					Domains:    types.Domains{types.Domain("example2.com")},
				}
				return types.Certificates{cert1, cert2}
			},
			domainsRequests: []*types.DomainRequest{{Requester: r, Domains: types.Domains{types.Domain("example2.com")}}},
			wantErr:         assert.NoError,
		},
		{
			name:  "SuccessNewCertWithIdentifierExist",
			state: &types.State{Certificates: types.Certificates{cert1}},
			wantFunc: func() types.Certificates {
				cert1 := &types.Certificate{
					Identifier: "example.com-0",
					Main:       "example.com",
					Domains:    types.Domains{types.Domain("example.com")},
				}
				cert2 := &types.Certificate{
					Identifier: "example.com-1",
					Main:       "example.com",
					Domains:    types.Domains{types.Domain("example.com"), types.Domain("example2.com")},
				}
				return types.Certificates{cert1, cert2}
			},
			domainsRequests: []*types.DomainRequest{{Requester: r, Domains: types.Domains{types.Domain("example.com"), types.Domain("example2.com")}}},

			wantErr: assert.NoError,
		},
		{
			name:  "SuccessNewCertWithMultipleSameRequest",
			state: &types.State{Certificates: types.Certificates{}},
			wantFunc: func() types.Certificates {
				return types.Certificates{cert1}
			},
			domainsRequests: []*types.DomainRequest{
				{Requester: r, Domains: types.Domains{types.Domain("example.com")}},
				{Requester: r, Domains: types.Domains{types.Domain("example.com")}},
			},

			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CertifierManager{}
			want := tt.wantFunc()
			cm.MatchingRequests(ctx, tt.state, tt.domainsRequests)

			for _, cert := range want {
				found := false
				for _, certificateState := range tt.state.Certificates {
					if reflect.DeepEqual(*cert, *certificateState) {
						found = true
					}
				}
				if !found {
					t.Errorf("cert not found %v", cert)
				}
			}
		})
	}
}

func TestCertifierManager_ObtainCertificates(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.Acme.RenewPeriod = time.Hour
	ctx.Config.Acme.DelayFailed = time.Hour
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	resolver := mockTypes.NewMockResolver(ctrl)
	fakeNow := time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC)
	tests := []struct {
		name      string
		state     *types.State
		mockFunc  func(resolver *mockTypes.MockResolver)
		checkFunc func(t *testing.T, state *types.State)
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "SuccessNothing",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}, Certificate: []byte("cert"), Key: []byte("key"), ExpirationDate: time.Now().Add(time.Hour * 2)},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resolver.EXPECT().ID().AnyTimes().Return(types.DefaultKey)
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
			},
			checkFunc: func(t *testing.T, state *types.State) {},
			wantErr:   assert.NoError,
		},
		{
			name: "SuccessObtainCertificate",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resource := &certificate.Resource{PrivateKey: []byte("key"), Certificate: []byte(certPemResponseMock)}
				resolver.EXPECT().ID().AnyTimes().Return(types.DefaultKey)
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
				resolver.EXPECT().Obtain(gomock.Any()).Times(1).Return(resource, nil)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.NotEmpty(t, cert.Key)
				assert.NotEmpty(t, cert.Certificate)
				assert.NotEmpty(t, cert.ExpirationDate)
			},
			wantErr: assert.NoError,
		},
		{
			name: "SuccessRenewCertificate",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}, Certificate: []byte("cert"), Key: []byte("key"), ExpirationDate: time.Now().Add(ctx.Config.Acme.RenewPeriod / 2)},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resource := &certificate.Resource{PrivateKey: []byte("key"), Certificate: []byte(certPemResponseMock)}
				resolver.EXPECT().ID().AnyTimes().Return(types.DefaultKey)
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
				resolver.EXPECT().RenewWithOptions(gomock.Any(), gomock.Any()).Times(1).Return(resource, nil)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.NotEmpty(t, cert.Key)
				assert.NotEmpty(t, cert.Certificate)
				assert.NotEmpty(t, cert.ExpirationDate)
			},
			wantErr: assert.NoError,
		},
		{
			name: "SuccessRenewCertificateWithMaxAttempt",
			state: &types.State{
				Certificates: types.Certificates{
					{
						Identifier:      "foo",
						Main:            "example.com",
						Domains:         types.Domains{types.Domain("example.com")},
						Certificate:     []byte("cert"),
						Key:             []byte("key"),
						ExpirationDate:  time.Now().Add(time.Hour * -2),
						ObtainFailDate:  fakeNow.Add(time.Hour * -2),
						ObtainFailCount: 3,
					},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resource := &certificate.Resource{PrivateKey: []byte("newKey"), Certificate: []byte(certPemResponseMock)}
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
				resolver.EXPECT().RenewWithOptions(gomock.Any(), gomock.Any()).Times(1).Return(resource, nil)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.Equal(t, "newKey", string(cert.Key))
				assert.NotEmpty(t, cert.Certificate)
				assert.NotEmpty(t, cert.ExpirationDate)
				assert.Equal(t, 0, cert.ObtainFailCount)
				assert.Equal(t, time.Time{}, cert.ObtainFailDate)
			},
			wantErr: assert.NoError,
		},
		{
			name: "SuccessNoRenewCertificate",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}, Certificate: []byte("cert"), Key: []byte("key"), ExpirationDate: time.Now().Add(time.Minute * 90)},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.NotEmpty(t, cert.Key)
				assert.NotEmpty(t, cert.Certificate)
				assert.NotEmpty(t, cert.ExpirationDate)
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailedObtainCertificateWithWildcard",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "*.example.com", Domains: types.Domains{types.Domain("*.example.com")}},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.Equal(t, 1, cert.ObtainFailCount)
				assert.Equal(t, fakeNow, cert.ObtainFailDate)
			},
			wantErr: assert.Error,
		},
		{
			name: "FailedRenewCertificate",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}, Certificate: []byte("cert"), Key: []byte("key"), ExpirationDate: time.Now().Add(time.Hour * -1)},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				//resolver.EXPECT().Match(gomock.Any()).Times(1).Return(true)
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
				resolver.EXPECT().RenewWithOptions(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("error"))
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.Equal(t, 1, cert.ObtainFailCount)
				assert.Equal(t, fakeNow, cert.ObtainFailDate)
			},
			wantErr: assert.Error,
		},
		{
			name: "SuccessSkipRenewCertificateWhenMaxAttempt",
			state: &types.State{
				Certificates: types.Certificates{
					{
						Identifier:      "foo",
						Main:            "example.com",
						Domains:         types.Domains{types.Domain("example.com")},
						Certificate:     []byte("cert"),
						Key:             []byte("key"),
						ExpirationDate:  time.Now().Add(time.Hour * -1),
						ObtainFailDate:  fakeNow.Add(time.Hour),
						ObtainFailCount: 3,
					},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
			},
			checkFunc: func(t *testing.T, state *types.State) {
				assert.Len(t, state.Certificates, 1)
				cert := state.Certificates[0]
				assert.Equal(t, 3, cert.ObtainFailCount)
				assert.Equal(t, fakeNow.Add(time.Hour), cert.ObtainFailDate)
			},
			wantErr: assert.NoError,
		},
		{
			name: "FailedDecodeCertificate",
			state: &types.State{
				Certificates: types.Certificates{
					{Identifier: "foo", Main: "example.com", Domains: types.Domains{types.Domain("example.com")}},
				},
			},
			mockFunc: func(resolver *mockTypes.MockResolver) {
				resource := &certificate.Resource{PrivateKey: []byte("key"), Certificate: []byte("wrong")}
				resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
				resolver.EXPECT().Obtain(gomock.Any()).Times(1).Return(resource, nil)
			},
			checkFunc: func(t *testing.T, state *types.State) {},
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.mockFunc(resolver)
			resolvers := types.Resolvers{types.DefaultKey: resolver}
			fakeClock := clockwork.NewFakeClockAt(fakeNow)
			cm := &CertifierManager{
				ephemeralID: "id",
				resolvers:   resolvers,
				clock:       fakeClock,
			}
			err := cm.ObtainCertificates(ctx, tt.state)
			tt.wantErr(t, err.ErrorOrNil(), fmt.Sprintf("ObtainCertificates(%v, %v, resolvers)", ctx, tt.state))
			tt.checkFunc(t, tt.state)
		})
	}
}

func TestCertifierManager_Run_SuccessWithNewAccount(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().Register(gomock.Any()).Times(1).Return(&registration.Resource{}, nil)
	resolvers := types.Resolvers{types.DefaultKey: resolver}

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: nil}, nil)
	storage.EXPECT().Save(gomock.Any()).Times(2).Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		resolvers:    resolvers,
	}
	err := cm.Run(ctx)
	assert.NoError(t, err)
}

func TestCertifierManager_Run_SuccessWithExistingAccount(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: account}, nil)
	storage.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err := cm.Run(ctx)
	assert.NoError(t, err)
}

func TestCertifierManager_Run_SuccessHasLock(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().Times(1).Return(&types.State{}, nil)
	err := ctx.Cache.Set(context.Background(), CacheProcessLockKey, "other")
	assert.NoError(t, err)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err = cm.Run(ctx)
	assert.NoError(t, err)
}

func TestCertifierManager_Run_FailedObtainLock(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := mockTypes.NewMockCache[string](ctrl)
	ctx.Cache = cache
	cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("error"))

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().Times(1).Return(&types.State{}, nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err := cm.Run(ctx)
	assert.Error(t, err)
}

func TestCertifierManager_Run_FailedReleaseLock(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := mockTypes.NewMockCache[string](ctrl)
	ctx.Cache = cache
	gomock.InOrder(
		cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", &store.NotFound{}),
		cache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil),
		cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("id", nil),
		cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error")),
	)
	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: account}, nil)
	storage.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err := cm.Run(ctx)
	assert.NoError(t, err)
}

func TestCertifierManager_Run_FailedLoadState(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)

	storage.EXPECT().Load().Times(1).Return(nil, errors.New("error"))
	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err := cm.Run(ctx)
	assert.Error(t, err)
}

func TestCertifierManager_Run_FailedCreateResolvers(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: account}, nil)
	ctx.Config.Acme.Resolvers["test"] = config.ResolverConfig{Type: "test"}
	dns.TypeDnsProviderMapping["test"] = func(ctx *appCtx.ServerContext, id string, config map[string]interface{}) (acme.Challenge, error) {
		return nil, errors.New("error")
	}
	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
	}
	err := cm.Run(ctx)
	assert.Error(t, err)
}

func TestCertifierManager_Run_FailRegisterAcmeAccount(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().Register(gomock.Any()).Times(1).Return(nil, errors.New("error"))
	resolvers := types.Resolvers{types.DefaultKey: resolver}

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: nil}, nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		resolvers:    resolvers,
	}
	err := cm.Run(ctx)
	assert.Error(t, err)
}

func TestCertifierManager_Run_FailFetchRequests(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	resolver := mockTypes.NewMockResolver(ctrl)
	resolvers := types.Resolvers{types.DefaultKey: resolver}

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: account, Certificates: types.Certificates{}}, nil)
	storage.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	requester := mockTypes.NewMockRequester(ctrl)
	requester.EXPECT().ID().Times(1).Return("foo")
	requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{}, errors.New("error"))

	ctx.Requesters = types.Requesters{"foo": requester}
	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		resolvers:    resolvers,
	}
	err := cm.Run(ctx)
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "unable to fetch requests")
}

func TestCertifierManager_Run_FailObtainCertificates(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().ID().Times(2).Return(types.DefaultKey)
	resolver.EXPECT().TypeChallenge().Times(1).Return(typesAcme.TypeHTTP01)
	resolver.EXPECT().Obtain(gomock.Any()).Times(1).Return(&certificate.Resource{}, nil)
	resolvers := types.Resolvers{types.DefaultKey: resolver}

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().Times(1).Return(&types.State{Account: account, Certificates: types.Certificates{{Identifier: "foo"}}}, nil)
	storage.EXPECT().Save(gomock.Any()).AnyTimes().Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())
	fakeClock := clockwork.NewFakeClockAt(time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC))

	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		resolvers:    resolvers,
		clock:        fakeClock,
	}
	err := cm.Run(ctx)
	assert.Nil(t, err)
}

func TestCertifierManager_Start(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctx.LogLevel.Set(slog.LevelDebug)
	ctx.Config.Interval = time.Millisecond * 100
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().AnyTimes().Return(&types.State{Account: account}, nil)
	storage.EXPECT().Save(gomock.Any()).AnyTimes().Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())

	fakeClock := clockwork.NewFakeClockAt(time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC))
	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		clock:        fakeClock,
	}

	go func() {
		err := cm.Start(ctx)
		assert.NoError(t, err)
	}()
	fakeClock.BlockUntil(1)
	fakeClock.Advance(ctx.Config.Interval)
	time.Sleep(ctx.Config.Interval)
	ctx.Cancel()
	assert.Contains(t, b.String(), "tick received")
	assert.Contains(t, b.String(), "stop asked by app, exiting")
}

func TestCertifierManager_StartFailFirstTick(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctx.LogLevel.Set(slog.LevelDebug)
	ctx.Config.Interval = time.Millisecond * 500 * 2
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	account, _ := acme.NewAccount("foo@bar.com")
	account.Registration = &registration.Resource{}
	storage.EXPECT().Load().AnyTimes().Return(nil, errors.New("error"))
	fakeClock := clockwork.NewFakeClockAt(time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC))
	cm := &CertifierManager{
		ephemeralID:  "id",
		stateStorage: storage,
		clock:        fakeClock,
	}
	go func() {
		err := cm.Start(ctx)
		assert.NoError(t, err)
	}()
	fakeClock.BlockUntil(1)
	fakeClock.Advance(ctx.Config.Interval)
	time.Sleep(50 * time.Millisecond)
	ctx.Cancel()
	assert.Contains(t, b.String(), "failed to load state")
}

func TestCertifierManager_initMetrics_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cert1 := &types.Certificate{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}
	metricsRegistry := mockPrometheus.NewMockRegistry(ctrl)

	gomock.InOrder(
		metricsRegistry.EXPECT().RegisterNewCertificateMetrics(gomock.Any()).Times(1),
		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(runCountMetric),
		metricsRegistry.EXPECT().MustAddCounter(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(fetchErrorMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(obtainCertErrorMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),
	)
	ctx.MetricsRegister = metricsRegistry

	cm := &CertifierManager{metricsInit: false}
	cm.initMetrics(ctx, &types.State{Certificates: types.Certificates{cert1}})
}

func TestCertifierManager_initMetrics_SuccessAlreadyInit(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	cm := &CertifierManager{metricsInit: true}
	cm.initMetrics(ctx, &types.State{})
}

func TestCertifierManager_CleanUnusedCertificates(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.UnusedRetentionDuration = time.Minute
	fakeNow := time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC)
	tests := []struct {
		name string

		certificates    types.Certificates
		domainsRequests []*types.DomainRequest
		want            types.Certificates
	}{
		{
			name:            "SuccessNoting",
			certificates:    types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}},
			domainsRequests: []*types.DomainRequest{{Domains: types.Domains{"example.com"}}},
			want:            types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}},
		},
		{
			name:            "SuccessUnusedCertificate",
			certificates:    types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}},
			domainsRequests: []*types.DomainRequest{},
			want:            types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}, UnusedAt: fakeNow}},
		},
		{
			name: "SuccessRemoveCertificate",
			certificates: types.Certificates{
				{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}},
				{Identifier: "bar", Main: "example2.com", Domains: types.Domains{"example2.com"}, UnusedAt: fakeNow.Add(time.Minute * -2)},
			},
			domainsRequests: []*types.DomainRequest{{Domains: types.Domains{"example.com"}}},
			want:            types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CertifierManager{
				clock: clockwork.NewFakeClockAt(fakeNow),
			}
			assert.Equalf(t, tt.want, cm.CleanUnusedCertificates(ctx, tt.certificates, tt.domainsRequests), "CleanUnusedCertificates(%v, %v, %v)", ctx, tt.certificates, tt.domainsRequests)
		})
	}
}

func TestCertifierManager_MarkCertificatesAsReused(t *testing.T) {

	tests := []struct {
		name            string
		certificates    types.Certificates
		domainsRequests []*types.DomainRequest
	}{
		{
			name:            "SuccessNoting",
			certificates:    types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}, UnusedAt: time.Time{}}},
			domainsRequests: []*types.DomainRequest{{Domains: types.Domains{"example.com"}}},
		},
		{
			name:            "SuccessReusedCertificate",
			certificates:    types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}, UnusedAt: time.Now()}},
			domainsRequests: []*types.DomainRequest{{Domains: types.Domains{"example.com"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CertifierManager{}
			cm.MarkCertificatesAsReused(tt.certificates, tt.domainsRequests)
			for _, c := range tt.certificates {
				assert.Equal(t, time.Time{}, c.UnusedAt)
			}
		})
	}
}

func TestCertifierManager_MarkCertificatesAsReused_StillNotUsed(t *testing.T) {
	now := time.Now()
	certificates := types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}, UnusedAt: now}}
	domainsRequests := []*types.DomainRequest{}
	cm := &CertifierManager{}
	cm.MarkCertificatesAsReused(certificates, domainsRequests)
	for _, c := range certificates {
		assert.Equal(t, now, c.UnusedAt)
	}
}
