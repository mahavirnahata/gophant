# Scheduler

Gophant includes a lightweight scheduler for periodic jobs.

## Create a schedule file

```bash
gophant make:schedule
```

This generates `schedule.go` that registers tasks using `gophant.RegisterSchedule`.

## Example

```go
func init() {
	gophant.RegisterSchedule(func(s *scheduler.Scheduler) {
		s.Every(10*time.Second, func() error {
			// task
			return nil
		})
	})
}
```

## How to run

Run once:

```bash
gophant schedule:run
```

Run continuously:

```bash
gophant schedule:work --interval 5
```

## Cron expressions (optional)

Use the cron helper:

```go
c := scheduler.NewCron()
_ = c.Add("@every 1m", func() error { return nil })

c.Start()
```
