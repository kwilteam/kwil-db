package chain

type Subscription interface {
	Unsubscribe()
	Err() <-chan error
}

type BlockSubscription interface {
	Unsubscribe()
	Err() <-chan error
	Blocks() <-chan int64
}
