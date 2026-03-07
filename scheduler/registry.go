package scheduler

var registrars []func(*Scheduler)

// RegisterSchedule registers a schedule builder.
func RegisterSchedule(fn func(*Scheduler)) {
	registrars = append(registrars, fn)
}

func ApplyRegisteredSchedules(s *Scheduler) {
	for _, fn := range registrars {
		fn(s)
	}
}
