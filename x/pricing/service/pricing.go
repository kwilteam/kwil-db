package service

import (
	"fmt"
	"kwil/x/pricing"
	"kwil/x/pricing/entity"
	"math/big"
)

type Service interface {
	EstimatePrice(request *entity.EstimateRequest) (*entity.EstimateResponse, error)
	GetPrice(requestType pricing.PricingRequestType) (*big.Int, error)
}

type pricingService struct {
}

func NewService() *pricingService {
	return &pricingService{}
}

func (p *pricingService) EstimatePrice(request *entity.EstimateRequest) (*entity.EstimateResponse, error) {
	// for now, we will just determine the request type and return a fixed price

	req := request.GetRequest()
	var price string
	switch req.(type) {
	case *entity.EstimateRequest_Deploy:
		price = entity.EstimatePrice(pricing.Deploy)
	case *entity.EstimateRequest_Delete:
		price = entity.EstimatePrice(pricing.Delete)
	case *entity.EstimateRequest_Query:
		price = entity.EstimatePrice(pricing.Query)
	}

	return &entity.EstimateResponse{
		Price: price,
	}, nil
}

func (p *pricingService) GetPrice(requestType pricing.PricingRequestType) (*big.Int, error) {
	bi, ok := new(big.Int).SetString(entity.GetPrice(requestType), 10)
	if !ok {
		return nil, fmt.Errorf("could not convert price to big.Int")
	}

	return bi, nil
}
