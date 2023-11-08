package types

type Subscription interface {
	Err() <-chan error
	Unsubscribe()
}
