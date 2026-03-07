package queue

type Job func() error

type Queue interface {
	Push(job Job) error
	Run() error
	Close() error
}
