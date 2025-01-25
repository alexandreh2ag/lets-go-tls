package os

import (
	"os"
	"os/user"
	"strconv"
)

func GetUserUID(username string) int {
	userInfo, err := user.Lookup(username)

	if err == nil && userInfo != nil {
		uid, errConv := strconv.Atoi(userInfo.Uid)
		if errConv == nil {
			return uid
		}
	}

	return os.Getuid()
}

func GetGroupUID(group string) int {
	groupInfo, err := user.Lookup(group)

	if err == nil && groupInfo != nil {
		gid, errConv := strconv.Atoi(groupInfo.Gid)
		if errConv == nil {
			return gid
		}
	}

	return os.Getgid()
}
