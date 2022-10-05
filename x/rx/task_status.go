package rx

const _VALUE = uint32(1)
const _ERROR = uint32(2)
const _CANCELLED = uint32(4)
const _FN = uint32(8)
const _LOCKED = uint32(16)
const _ASYNC_CONTINUATIONS = uint32(32)
const _INCOMPLETE_ORIGIN = uint32(64)

const _CANCELLED_OR_ERROR = _ERROR | _CANCELLED
const _DONE = _VALUE | _ERROR | _CANCELLED
const _ANY_HANDLER = _FN | _INCOMPLETE_ORIGIN

var _closedChan <-chan Void

func init() {
	ch := make(chan Void)
	close(ch)
	_closedChan = ch
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
