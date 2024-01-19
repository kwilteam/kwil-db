package pg

// This file declares a package-level logger since passing loggers around and
// creating logger fields is annoying sometimes, particularly in this package.

import klog "github.com/kwilteam/kwil-db/core/log"

var logger = klog.NewNoOp()

func UseLogger(log klog.Logger) {
	logger = log
}
