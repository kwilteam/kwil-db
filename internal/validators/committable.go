package validators

type Committable interface {
	SetIDFunc(func() ([]byte, error))
	Skip() bool
}
