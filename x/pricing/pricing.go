package pricing

type PricingRequestType int

const (
	Deploy PricingRequestType = iota
	Delete
	Query
)
