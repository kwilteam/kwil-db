package async

const _VALUE = uint32(1)
const _ERROR = uint32(2)
const _CANCELLED = uint32(4)
const _FN = uint32(8)
const _LOCKED = uint32(16)
const _ASYNC_CONTINUATIONS = uint32(32)
const _CANCELLED_OR_ERROR = _ERROR | _CANCELLED
const _DONE = _VALUE | _ERROR | _CANCELLED

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

func isDone(current uint32) bool {
	return current&_DONE != 0
}
