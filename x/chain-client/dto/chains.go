package dto

type ChainType int // this is used to indicate the chain ID.
// I am using this instead of typicaly ChainIDs beecause we will want to support non-EVM chains (which have no chain ID)

const (
	UNKNWON_CHAIN ChainType = iota
	ETHEREUM
	GOERLI
)
