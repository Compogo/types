package time

import (
	"context"
	"sync/atomic"
	"time"
)

type Timer struct {
	C chan time.Time

	timer *time.Timer
	isRun atomic.Bool

	cancelFunc context.CancelFunc
}

func NewTimer(ctx context.Context, d time.Duration) *Timer {
	timer := &Timer{
		C:     make(chan time.Time, 1),
		timer: time.NewTimer(d),
	}
	timer.isRun.Store(true)
	timer.chanProcess(ctx)

	return timer
}

func (timer *Timer) Stop() bool {
	defer timer.isRun.Store(false)
	defer func() {
		if timer.cancelFunc != nil {
			timer.cancelFunc()
		}
	}()

	return timer.timer.Stop()
}

func (timer *Timer) Reset(ctx context.Context, d time.Duration) bool {
	defer timer.isRun.Store(true)
	defer timer.chanProcess(ctx)

	return timer.timer.Reset(d)
}

func (timer *Timer) IsRun() bool {
	return timer.isRun.Load()
}

func (timer *Timer) chanProcess(ctx context.Context) {
	if timer.cancelFunc != nil {
		timer.cancelFunc()
	}

	ctx, timer.cancelFunc = context.WithCancel(ctx)
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case val := <-timer.timer.C:
			defer timer.isRun.Store(false)
			timer.C <- val
			return
		}
	}(ctx)
}
