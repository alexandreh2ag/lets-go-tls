package cli

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Version_ExecuteCommand(t *testing.T) {
	cmd := GetVersionCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	_ = cmd.Execute()

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "develop-SNAPSHOT\n", string(out))
}
