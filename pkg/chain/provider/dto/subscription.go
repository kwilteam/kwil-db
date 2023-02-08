package dto

type Subscription interface {
	Unsubscribe()
	Err() <-chan error
}
