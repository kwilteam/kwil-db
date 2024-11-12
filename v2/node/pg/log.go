package pg

// This file declares a package-level logger since passing loggers around and
// creating logger fields is annoying sometimes, particularly in this package.

import "kwil/log"

var logger = log.DiscardLogger

func UseLogger(log log.Logger) {
	logger = log
}
