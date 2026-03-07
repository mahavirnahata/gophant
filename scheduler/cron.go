package scheduler

import "github.com/robfig/cron/v3"

type Cron struct {
	c *cron.Cron
}

func NewCron() *Cron {
	return &Cron{c: cron.New()}
}

func (c *Cron) Add(spec string, task Task) error {
	_, err := c.c.AddFunc(spec, func() { _ = task() })
	return err
}

func (c *Cron) Start() {
	c.c.Start()
}

func (c *Cron) Stop() {
	c.c.Stop()
}
