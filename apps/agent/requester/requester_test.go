package requester

import (
	"errors"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFetchRequests_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	requester := mockTypes.NewMockRequester(ctrl)
	want1 := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("foo.com")}},
		{Requester: requester, Domains: types.Domains{types.Domain("bar.com")}},
	}
	want2 := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("bar.foo.com"), types.Domain("foo.bar.com")}},
	}
	want := append(want1, want2...)
	requesters := types.Requesters{"foo": requester, "bar": requester}
	gomock.InOrder(
		requester.EXPECT().Fetch().Times(1).Return(want1, nil),
		requester.EXPECT().Fetch().Times(1).Return(want2, nil),
	)

	got, err := FetchRequests(requesters)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestFetchRequests_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	requester := mockTypes.NewMockRequester(ctrl)
	want := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("foo.com")}},
		{Requester: requester, Domains: types.Domains{types.Domain("bar.com")}},
	}
	requesters := types.Requesters{"foo": requester, "bar": requester}
	gomock.InOrder(
		requester.EXPECT().Fetch().Times(1).Return(want, nil),
		requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{}, errors.New("fail")),
		requester.EXPECT().ID().Times(1).Return("bar"),
	)

	got, err := FetchRequests(requesters)
	assert.Error(t, err)
	assert.Equal(t, want, got)
}
