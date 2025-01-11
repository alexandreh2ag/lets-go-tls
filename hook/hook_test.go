package hook

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestNewManagerHook(t *testing.T) {
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)
	opts := &slog.HandlerOptions{AddSource: false, Level: level}
	logger := slog.New(slog.NewTextHandler(io.Discard, opts))

	got := NewManagerHook(logger)
	assert.NotNil(t, got)
}

func TestManagerHook_GetHookChan(t *testing.T) {
	m := &ManagerHook{hookChan: make(chan *Hook)}
	assert.NotNil(t, m.GetHookChan())
}

func TestManagerHook_RunHook(t *testing.T) {

	tests := []struct {
		name    string
		hook    *Hook
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "success",
			hook:    &Hook{Cmd: "echo 1"},
			wantErr: assert.NoError,
		},
		{
			name:    "successWithBash",
			hook:    &Hook{Cmd: "bash -c 'echo 1 && echo 2'", Timeout: time.Second},
			wantErr: assert.NoError,
		},
		{
			name:    "faild",
			hook:    &Hook{Cmd: "cmdfailnotexit", Timeout: time.Second},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ManagerHook{}
			tt.wantErr(t, m.RunHook(tt.hook), fmt.Sprintf("RunHook(%v)", tt.hook))
		})
	}
}

func Test_splitCommand(t *testing.T) {

	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "SuccessSimpleCommand",
			command: "echo",
			want:    []string{"echo"},
		},
		{
			name:    "SuccessComplexCommand",
			command: "echo test",
			want:    []string{"echo", "test"},
		},
		{
			name:    "SuccessComplexCommandWithDoubleQuote",
			command: "echo \"test\"",
			want:    []string{"echo", "\"test\""},
		},
		{
			name:    "SuccessComplexCommandWithQuote",
			command: "echo 'test'",
			want:    []string{"echo", "'test'"},
		},
		{
			name:    "SuccessComplexCommandWithQuoteAndDoubleQuote",
			command: "echo '\"test\"'",
			want:    []string{"echo", "'\"test\"'"},
		},
		{
			name:    "SuccessComplexCommandWithDoubleQuoteAndSpace",
			command: "echo ' test '",
			want:    []string{"echo", " test "},
		},
		{
			name:    "SuccessComplexCommandWithQuoteAndSpace",
			command: "echo ' test '",
			want:    []string{"echo", " test "},
		},
		{
			name:    "SuccessBashWrappedCommand",
			command: "bash -c 'echo test'",
			want:    []string{"bash", "-c", "echo test"},
		},
		{
			name:    "SuccessBashWrappedCommandAndSpace",
			command: "bash -c ' echo test '",
			want:    []string{"bash", "-c", " echo test "},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, splitCommand(tt.command), "splitCommand(%v)", tt.command)
		})
	}
}

func TestManagerHook_RunHooks1(t *testing.T) {

	tests := []struct {
		name     string
		hooks    []*Hook
		contains string
	}{
		{
			name: "success",
			hooks: []*Hook{
				{Cmd: "echo 1", Timeout: time.Second},
			},
			contains: "",
		},
		{
			name: "fail",
			hooks: []*Hook{
				{Cmd: "cmdfailnotexit", Timeout: time.Second},
			},
			contains: "failed to run hook cmdfailnotexit: executing hook 'cmdfailnotexit'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			level := &slog.LevelVar{}
			level.Set(slog.LevelInfo)
			opts := &slog.HandlerOptions{AddSource: false, Level: level}
			logger := slog.New(slog.NewTextHandler(b, opts))
			m := &ManagerHook{
				hooks:  tt.hooks,
				logger: logger,
			}
			m.RunHooks()
			assert.Contains(t, b.String(), tt.contains)
		})
	}
}

func TestManagerHook_Start(t *testing.T) {
	hookChan := make(chan *Hook)
	m := &ManagerHook{
		hookChan: hookChan,
	}
	go m.Start()
	h := &Hook{Cmd: "echo 1"}
	hookChan <- h
	assert.Equal(t, []*Hook{h}, m.hooks)
}
