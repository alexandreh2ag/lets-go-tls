package version

import (
	"fmt"
)

var (
	Commit  = "SNAPSHOT"
	Version = "develop"
)

func GetFormattedVersion() string {
	var res string
	// Commit var can be updated by build args
	if Commit != "" {
		res = fmt.Sprintf("%s-%s", Version, Commit)
	} else {
		res = Version
	}
	return res
}
