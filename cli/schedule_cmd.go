package cli

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/scheduler"
)

func ScheduleRunOnce() error {
	s := scheduler.New()
	gophant.ApplyRegisteredSchedules(s)
	s.RunOnce()
	return nil
}

func ScheduleRunLoop() error {
	s := scheduler.New()
	gophant.ApplyRegisteredSchedules(s)

	stop := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		close(stop)
	}()

	s.Run(time.Second, stop)
	return nil
}
