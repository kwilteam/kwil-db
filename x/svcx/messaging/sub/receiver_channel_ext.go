package sub

//
//import (
//	"kwil/x"
//	"kwil/x/svcx/messaging/mx"
//	"kwil/x/syncx"
//)
//
//type receiver_channel struct {
//	topic       string
//	partitionId mx.PartitionId
//	out         syncx.Chan[MessageIterator]
//	done        chan x.Void
//}
//
//func new_receiver_channel(topic string, partitionId mx.PartitionId) receiver_channel {
//	return receiver_channel{
//		topic:       topic,
//		partitionId: partitionId,
//		out:         syncx.NewChanBuffered[MessageIterator](1),
//		done:        make(chan x.Void),
//	}
//}
//
//// chan MessageIterator
//func (c *receiver_channel) push(MessageIterator) string {
//	// put into internal channel
//	// use an event loop to push to out channel?
//	panic("not implemented")
//}
//
//func (c *receiver_channel) Topic() string {
//	return c.topic
//}
//
//func (c *receiver_channel) PartitionId() mx.PartitionId {
//	return c.partitionId
//}
//
//func (c *receiver_channel) OnReceive() <-chan MessageIterator {
//	return c.out.Read()
//}
//
//func (c *receiver_channel) OnStop() <-chan x.Void {
//	return c.done
//}
//
//func (c *receiver_channel) Stop() {
//	close(c.done)
//}
//
//func (c *receiver_channel) close() {
//	c.out.Close()
//}
