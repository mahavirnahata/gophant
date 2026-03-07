package cli

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahavirnahata/gophant/scheduler"
)

type ScheduleFunc func(*scheduler.Scheduler)

type ScheduleOptions struct {
	Interval time.Duration
}

func ScheduleRun(fn ScheduleFunc, opts ScheduleOptions) {
	s := scheduler.New()
	fn(s)

	stop := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		close(stop)
	}()

	interval := opts.Interval
	if interval <= 0 {
		interval = time.Second
	}

	s.Run(interval, stop)
}
