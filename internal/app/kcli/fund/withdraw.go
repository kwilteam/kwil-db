package fund

/*
func withdrawCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraws funds from the funding pool",
		Long:  `"withdraw" withdraws funds from the funding pool.`,
		Args:  cobra.ExactArgs(1),
		//RunE: func(cmd *cobra.Command, args []string) error {
		// send request

		//return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
		/*client := apipb.NewKwilServiceClient(cc)
						c, err := chain.NewClientV(viper.GetViper())
						if err != nil {
							return fmt.Errorf("error creating chain client: %w", err)
						}

						// get balance
						balance, err := c.GetDepositBalance()
						if err != nil {
							return fmt.Errorf("error getting deposit balance: %w", err)
						}

						amount, ok := new(big.Int).SetString(args[0], 10)
						if !ok {
							return errors.New("could not convert amount to big int")
						}

						if balance.Cmp(amount) < 0 {
							return fmt.Errorf("insufficient funds: %s of %s", amount, balance)
						}

						addr := c.Address.String()

						// now we will send a request to "/api/v0/wallets/withdraw"
						n := common.GenerateNonce(10)

						// generate id
						id := crypto.Sha384Str([]byte(amount.String() + n + addr))

						// sign it
						sig, err := crypto.Sign([]byte(id), c.Wallet)
						if err != nil {
							return fmt.Errorf("error signing request: %w", err)
						}

						fmt.Println("THIS IS NOT DONE ThIS DOES NOT WORK YET")

						wdr := apipb.WithdrawalRequest{
							Tx: &apipb.Tx{
								Id:        id,
								Payload:   []byte{},
								Fee:       amount.String(),
								Nonce:     n,
								Sender:    addr,
								Signature: sig,
							},
						}
						res, err := client.ReturnFunds(ctx, &wdr)
						if err != nil {
							fmt.Printf("%+v", &wdr)
							return fmt.Errorf("error sending request: %w", err)
						}

						fmt.Printf(`Withdrawal request sent.

			Amount Requested: %s
			Amount Returned:  %s
			Fee:              %s
			Correlation ID:   %s
			Tx Hash:          %s

		`, amount, res.Amount, res.Fee, res.CorrelationId, res.Txid)
		//return nil
		//})
		//},
	}

	return cmd
}
*/
