package version

import (
	"testing"
)

func TestGetFormattedVersion(t *testing.T) {
	want := "develop-SNAPSHOT"
	if got := GetFormattedVersion(); got != want {
		t.Errorf("GetFormattedVersion() = %v, want %v", got, want)
	}
}

func TestGetFormattedVersionWhenCommitIsEmpty(t *testing.T) {
	want := "develop"
	Commit = ""
	if got := GetFormattedVersion(); got != want {
		t.Errorf("GetFormattedVersion() = %v, want %v", got, want)
	}
}
