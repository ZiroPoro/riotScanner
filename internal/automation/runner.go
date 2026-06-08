package automation

import (
	"fmt"
	"time"

	"github.com/edu/riotscanner/internal/adb"
)

type Runner struct {
	client  *adb.Client
	actions []TapAction
	pollMs  int
}

func NewRunner(client *adb.Client, actions []TapAction, pollIntervalMs int) *Runner {
	if pollIntervalMs <= 0 {
		pollIntervalMs = 500
	}
	return &Runner{
		client:  client,
		actions: actions,
		pollMs:  pollIntervalMs,
	}
}

func (r *Runner) RunOnce() (string, error) {
	raw, err := r.client.Screenshot()
	if err != nil {
		return "", err
	}

	img, err := LoadImage(raw)
	if err != nil {
		return "", err
	}

	for _, action := range r.actions {
		matcher := NewScreenMatcher(action.Probes)
		if matcher.MatchAll(img) {
			if err := r.client.Tap(action.TapX, action.TapY); err != nil {
				return action.Name, fmt.Errorf("tap %s: %w", action.Name, err)
			}
			return action.Name, nil
		}
	}
	return "", nil
}

func (r *Runner) RunLoop(maxIterations int, onMatch func(name string) error) error {
	for i := 0; maxIterations <= 0 || i < maxIterations; i++ {
		name, err := r.RunOnce()
		if err != nil {
			return err
		}
		if name != "" {
			if onMatch != nil {
				if err := onMatch(name); err != nil {
					return err
				}
			}
		}
		time.Sleep(time.Duration(r.pollMs) * time.Millisecond)
	}
	return nil
}
