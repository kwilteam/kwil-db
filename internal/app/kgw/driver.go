package kgw

// KgwDriver is a driver for the http client for integration tests
type KgwDriver struct {
}

func NewKgwDriver() *KgwDriver {
	return &KgwDriver{}
}
