package os

import (
	"github.com/stretchr/testify/assert"
	"os/user"
	"strconv"
	"testing"
)

func TestGetUserUID(t *testing.T) {
	userInfo, err := user.Current()
	assert.Nil(t, err)

	uid, errConv := strconv.Atoi(userInfo.Uid)
	assert.NoError(t, errConv)

	tests := []struct {
		name     string
		username string
		want     int
	}{
		{
			name:     "Get user uid",
			username: userInfo.Username,
			want:     uid,
		},
		{
			name:     "Get current user uid",
			username: "not-exist",
			want:     uid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserUID(tt.username); got != tt.want {
				t.Errorf("GetUserUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGroupUID(t *testing.T) {
	userInfo, err := user.Current()
	assert.Nil(t, err)

	groupInfo, errGroup := user.LookupGroupId(userInfo.Gid)
	assert.NoError(t, errGroup)

	gid, errConv := strconv.Atoi(userInfo.Gid)
	assert.NoError(t, errConv)

	tests := []struct {
		name  string
		group string
		want  int
	}{
		{
			name:  "Get group uid",
			group: groupInfo.Name,
			want:  gid,
		},
		{
			name:  "Get current group uid",
			group: "not-exist",
			want:  gid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGroupUID(tt.group); got != tt.want {
				t.Errorf("GetGroupUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
