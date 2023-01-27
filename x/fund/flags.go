package fund

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FundingPoolFlag = "funding-pool"
	FundingPoolEnv  = "KWIL_FUNDING_POOL"

	ValidatorAddressFlag = "validator-address"
	ValidatorAddressEnv  = "KWIL_VALIDATOR_ADDRESS"

	TokenAddressFlag = "token-address"
	TokenAddressEnv  = "KWIL_TOKEN_ADDRESS"
)

func BindFundFlags(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	fs.String(FundingPoolFlag, "", "funding pool")
	viper.BindPFlag(FundingPoolFlag, fs.Lookup(FundingPoolFlag))

	fs.String(ValidatorAddressFlag, "", "validator address")
	viper.BindPFlag(ValidatorAddressFlag, fs.Lookup(ValidatorAddressFlag))
}

func BindFundEnv(cmd *cobra.Command) {
	viper.BindEnv(FundingPoolFlag, FundingPoolEnv)
	viper.BindEnv(ValidatorAddressFlag, ValidatorAddressEnv)
}
