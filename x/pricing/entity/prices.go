package entity

import (
	"kwil/x/pricing"
)

// this is by no means a complete implementation of the pricing service

const (
	CREATE_PRICE = "1000000000000000000"
	DROP_PRICE   = "10000000000000"
	QUERY_PRICE  = "2000000000000000"
)

func EstimatePrice(p pricing.PricingRequestType) string {
	switch p {
	case pricing.Deploy:
		return CREATE_PRICE
	case pricing.Delete:
		return DROP_PRICE
	case pricing.Query:
		return QUERY_PRICE
	}
	return "0"
}

func GetPrice(p pricing.PricingRequestType) string {
	switch p {
	case pricing.Deploy:
		return CREATE_PRICE
	case pricing.Delete:
		return DROP_PRICE
	case pricing.Query:
		return QUERY_PRICE
	}
	return "0"
}
