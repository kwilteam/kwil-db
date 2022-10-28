package async

type OptPool struct {
	fn func(e *worker_pool)
}

func (o *OptPool) configure(e *worker_pool) {
	o.fn(e)
}

func WithErrorHandler(unhandledErrorHandler func(error)) *OptPool {
	return &OptPool{func(e *worker_pool) {
		e.fn = unhandledErrorHandler
	}}
}

func WithMaxQueueLength(queue_length int) *OptPool {
	return &OptPool{func(e *worker_pool) {
		e.queue_length = queue_length
	}}
}
