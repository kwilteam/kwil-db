package datasets

import "kwil/pkg/log"

type DatasetUseCaseOpt func(*DatasetUseCase)

func WithLogger(logger log.Logger) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.log = logger
	}
}
