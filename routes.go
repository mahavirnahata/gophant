package gophant

import "github.com/mahavirnahata/gophant/scheduler"

var routeRegistrars []func(*App)

// RegisterRoutes allows apps to register routes during init().
func RegisterRoutes(fn func(*App)) {
	routeRegistrars = append(routeRegistrars, fn)
}

func applyRegisteredRoutes(app *App) {
	for _, fn := range routeRegistrars {
		fn(app)
	}
}

// Schedule registration for scheduler magic.
var scheduleRegistrars []func(*scheduler.Scheduler)

// RegisterSchedule allows apps to register scheduled jobs during init().
func RegisterSchedule(fn func(*scheduler.Scheduler)) {
	scheduleRegistrars = append(scheduleRegistrars, fn)
}

func ApplyRegisteredSchedules(s *scheduler.Scheduler) {
	for _, fn := range scheduleRegistrars {
		fn(s)
	}
}
