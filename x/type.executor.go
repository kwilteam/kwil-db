package x

type Job Runnable

type Executor interface {
	Execute(Job)
}
