package sub

//
//import (
//	"context"
//	"kwil/x"
//	"kwil/x/messaging/mx"
//	"kwil/x/syncx"
//)
//
//type receiver_channel[T any] struct {
//	serdes      mx.Serdes[T]
//	topic       string
//	partitionId mx.PartitionId
//	out         syncx.Chan[MessageIterator[T]]
//	done        chan x.Void
//}
//
//func new_receiver_channel[T any](serdes mx.Serdes[T], topic string, partitionId mx.PartitionId) receiver_channel[T] {
//	return receiver_channel[T]{
//		serdes:      serdes,
//		topic:       topic,
//		partitionId: partitionId,
//		out:         syncx.NewChanBuffered[MessageIterator[T]](1),
//		done:        make(chan x.Void),
//	}
//}
//
//// Implement the ReceiverChannel interface and MessageIterator
//// MessageIterator will return false when close is called
////
//
//// chan MessageIterator[T]
//func (c *receiver_channel[T]) push(MessageIterator[T]) string {
//	// put into internal channel
//	// use an event loop to push to out channel?
//}
//
//func (c *receiver_channel[T]) Topic() string {
//	return c.topic
//}
//
//func (c *receiver_channel[T]) PartitionId() mx.PartitionId {
//	return c.partitionId
//}
//
//func (c *receiver_channel[T]) OnReceive() <-chan MessageIterator[T] {
//	return c.out.Read()
//}
//
//func (c *receiver_channel[T]) OnClosed() <-chan x.Void {
//	return c.done
//}
//
//func (c *receiver_channel[T]) Close() {
//	close(c.done)
//}
//
//func (c *receiver_channel[T]) CloseAndWait(ctx context.Context) error {
//	c.Close()
//	return nil
//}
//
//func (c *receiver_channel[T]) close() {
//	c.out.Close()
//}
