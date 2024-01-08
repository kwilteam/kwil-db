package chain

type Header struct {
	Hash   string
	Height int64
}

type Subscription interface {
	Unsubscribe()
	Err() <-chan error
}

type BlockSubscription interface {
	Unsubscribe()
	Err() <-chan error
	Blocks() <-chan int64
}
