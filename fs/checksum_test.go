package fs

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChecksum_ChecksumByPath_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"

	_ = afero.WriteFile(fs, path, []byte("test"), 0644)
	hash, err := checksum.ChecksumByPath(path)
	assert.NoError(t, err)
	assert.NotEqual(t, []byte{}, hash)
}

func TestChecksum_ChecksumByPath_FailReadFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"

	hash, err := checksum.ChecksumByPath(path)
	assert.Error(t, err)
	assert.Equal(t, []byte{}, hash)
}

func TestChecksum_MustChecksumByPath_SuccessFileExist(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"

	_ = afero.WriteFile(fs, path, []byte("test"), 0644)
	hash := checksum.MustChecksumByPath(path)
	assert.NotEqual(t, []byte{}, hash)
}

func TestChecksum_MustChecksumByPath_SuccessFileNotExist(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"

	hash := checksum.MustChecksumByPath(path)
	assert.Equal(t, []byte{}, hash)
}

func TestChecksum_ChecksumByContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)

	hash := checksum.ChecksumByContent([]byte("test"))
	assert.NotEqual(t, []byte{}, hash)
}

func TestChecksum_MustCompareContentWithPath_SuccessEqual(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"
	_ = afero.WriteFile(fs, path, []byte("test"), 0644)
	equal := checksum.MustCompareContentWithPath([]byte("test"), path)
	assert.True(t, equal)
}

func TestChecksum_MustCompareContentWithPath_SuccessNotEqual(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"
	_ = afero.WriteFile(fs, path, []byte("foo"), 0644)
	equal := checksum.MustCompareContentWithPath([]byte("test"), path)
	assert.False(t, equal)
}

func TestChecksum_MustCompareContentWithPath_SuccessNotEqualWithFileNotExist(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"

	equal := checksum.MustCompareContentWithPath([]byte("test"), path)
	assert.False(t, equal)
}

func TestChecksum_MustCompareContentWithPath_SuccessNotEqualWithEmptyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	checksum := NewChecksum(fs)
	path := "/app/test.txt"
	_ = afero.WriteFile(fs, path, []byte(""), 0644)
	equal := checksum.MustCompareContentWithPath([]byte("test"), path)
	assert.False(t, equal)
}
