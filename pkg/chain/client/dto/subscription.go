package dto

type Subscription interface {
	Unsubscribe()
	Err() <-chan error
	Blocks() <-chan int64
}
