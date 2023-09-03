package fs

import (
	"bytes"
	"crypto/sha256"
	"github.com/spf13/afero"
)

type Checksum struct {
	fs afero.Fs
}

func NewChecksum(fs afero.Fs) *Checksum {
	return &Checksum{fs: fs}
}

func (c *Checksum) ChecksumByPath(path string) ([]byte, error) {
	content, err := afero.ReadFile(c.fs, path)
	if err != nil {
		return []byte{}, err
	}
	return c.ChecksumByContent(content), nil
}

func (c *Checksum) MustChecksumByPath(path string) []byte {
	content, _ := afero.ReadFile(c.fs, path)

	if content == nil {
		return []byte{}
	}

	return c.ChecksumByContent(content)
}

func (c *Checksum) MustCompareContentWithPath(content []byte, path string) bool {
	hashPath := c.MustChecksumByPath(path)

	return bytes.Compare(c.ChecksumByContent(content), hashPath) == 0
}

func (c *Checksum) ChecksumByContent(content []byte) []byte {
	h := sha256.New()
	h.Write(content)

	return h.Sum(nil)
}
