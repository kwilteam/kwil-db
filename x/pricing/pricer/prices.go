package pricer

import (
	"kwil/x/pricing"
)

// this is by no means a complete implementation of the pricing service

const (
	CREATE_PRICE = "1000000000000000000"
	DROP_PRICE   = "10000000000000"
	QUERY_PRICE  = "2000000000000000"
)

func GetPrice(p pricing.PricingRequestType) string {
	switch p {
	case pricing.DEPLOY:
		return CREATE_PRICE
	case pricing.DROP:
		return DROP_PRICE
	case pricing.QUERY:
		return QUERY_PRICE
	}
	return "0"
}
