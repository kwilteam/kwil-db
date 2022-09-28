package rx

const _VALUE = uint32(1)                   //00000001
const _ERROR = uint32(2)                   //00000010
const _CANCELLED = uint32(4)               //00000100
const _FN = uint32(8)                      //00001000
const _LOCKED = uint32(16)                 //00010000
const _ASYNC_CONTINUATIONS = uint32(32)    //00100000
const _INCOMPLETE_ORIGIN = uint32(64)      //01000000
const _DONE_BLOCKING_HANDLER = uint32(128) //10000000
const _CANCELLED_OR_ERROR = _ERROR | _CANCELLED
const _DONE = _VALUE | _ERROR | _CANCELLED
const _ANY_HANDLER = _FN | _INCOMPLETE_ORIGIN

var _taskValue = _VALUE
var _taskValuePtr = &_taskValue
var _taskErrorPtr = &_taskValue

var _closedChan = make(chan struct{})

func init() {
	close(_closedChan)
}

func isAsync(status uint32) bool {
	return status&_ASYNC_CONTINUATIONS != 0
}

func isLocked(status uint32) bool {
	return status&_LOCKED != 0
}

func hasError(current uint32) bool {
	return current&_CANCELLED_OR_ERROR != 0
}

func isCancelled(current uint32) bool {
	return current&_CANCELLED != 0
}

func hasHandler(current uint32) bool {
	return current&_FN != 0
}

func hasAnyHandler(current uint32) bool {
	return current&_ANY_HANDLER != 0
}

func isDone(current uint32) bool {
	return current&_DONE != 0
}

func isCompletedOrigin(current uint32) bool {
	return current&_INCOMPLETE_ORIGIN != 0
}

func hasBlockingDoneHandler(current uint32) bool {
	return current&_DONE_BLOCKING_HANDLER != 0
}
