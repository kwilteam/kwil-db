package pricing

type PricingRequestType int

const (
	DEPLOY PricingRequestType = iota
	DROP
	QUERY
	WITHDRAW
)
