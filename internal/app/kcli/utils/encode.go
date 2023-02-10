package utils

import (
	"fmt"
	"kwil/pkg/databases/spec"

	"github.com/spf13/cobra"
)

func encodeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "encode",
		Short: "Encode is used to encode a given input type.",
		Long: `The Encode function encodes an input for a specific type.
The type is specified by the --type flag. The input is specified by the --input flag.
The output is printed to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// get type
			t, err := cmd.Flags().GetString("type")
			if err != nil {
				return fmt.Errorf("error getting type: %w", err)
			}

			// convert type
			typ, err := spec.DataTypeConversions.StringToKwilType(t)
			if err != nil {
				fmt.Printf("type must be one of: %v",
					[]string{"null", "string", "int32", "int64", "boolean"})
				return fmt.Errorf("error converting type: %w", err)
			}

			// get input
			i, err := cmd.Flags().GetString("input")
			if err != nil {
				return fmt.Errorf("error getting input: %w", err)
			}

			// encode
			val, err := spec.NewExplicit(i, typ)
			if err != nil {
				return fmt.Errorf("error encoding input: %w", err)
			}

			// print
			val.Print()

			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "The type to encode the input as.")
	cmd.MarkFlagRequired("type")
	cmd.Flags().StringP("input", "i", "", "The input to encode.")
	cmd.MarkFlagRequired("input")

	return cmd
}
