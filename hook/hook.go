package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ManagerHook struct {
	hookChan chan *Hook
	hooks    []*Hook
	mutex    sync.Mutex
	logger   *slog.Logger
}

func NewManagerHook(logger *slog.Logger) *ManagerHook {
	return &ManagerHook{
		hookChan: make(chan *Hook),
		hooks:    []*Hook{},
		logger:   logger,
	}
}

func (m *ManagerHook) Start() {
	for {
		select {
		case hook := <-m.hookChan:
			m.mutex.Lock()
			m.hooks = append(m.hooks, hook)
			m.mutex.Unlock()
		}
	}
}

func (m *ManagerHook) GetHookChan() chan<- *Hook {
	return m.hookChan
}

func (m *ManagerHook) deduplicate() []*Hook {
	hooks := []*Hook{}
	seen := make(map[string]bool)
	for _, hook := range m.hooks {
		if !seen[hook.Cmd] {
			seen[hook.Cmd] = true
			hooks = append(hooks, hook)
		}
	}
	return hooks
}

func (m *ManagerHook) RunHooks() {
	// wait all storage send hook
	time.Sleep(time.Second * 1)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger.Debug(fmt.Sprintf("manager hook: run hooks (%d)", len(m.hooks)))
	m.hooks = m.deduplicate()
	for _, hook := range m.hooks {
		err := m.RunHook(hook)
		if err != nil {
			m.logger.Error(fmt.Sprintf("failed to run hook %s: %s", hook.Cmd, err.Error()))
		}
	}
	m.hooks = []*Hook{}
}

func (m *ManagerHook) RunHook(hook *Hook) error {
	if hook.Timeout == 0 {
		hook.Timeout = time.Minute * 1
	}

	cmd, cancel := context.WithTimeout(context.Background(), hook.Timeout)
	defer cancel()

	parts := splitCommand(os.ExpandEnv(hook.Cmd))

	cmdCtx := exec.CommandContext(cmd, parts[0], parts[1:]...)

	output, err := cmdCtx.CombinedOutput()

	if err != nil {
		return fmt.Errorf("executing hook '%s': %s with output %s", hook.Cmd, err, output)
	}

	return nil
}

func splitCommand(command string) []string {
	split := strings.Split(command, " ")

	var result []string
	var inquote string
	var block string
	for _, i := range split {
		if inquote == "" {
			if (strings.HasPrefix(i, "'") || strings.HasPrefix(i, "\"")) && !(len(i) > 2 && (strings.HasSuffix(i, "'") || strings.HasSuffix(i, "\""))) {
				inquote = string(i[0])
				block = strings.TrimPrefix(i, inquote) + " "
			} else {
				result = append(result, i)
			}
		} else {
			if !strings.HasSuffix(i, inquote) {
				block += i + " "
			} else {
				block += strings.TrimSuffix(i, inquote)
				inquote = ""
				result = append(result, block)
				block = ""
			}
		}
	}
	return result
}

type Hook struct {
	Cmd     string        `mapstructure:"cmd" validate:"required"`
	Timeout time.Duration `mapstructure:"timeout"`
}
