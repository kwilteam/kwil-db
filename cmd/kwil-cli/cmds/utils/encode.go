package utils

import (
	"fmt"
	"kwil/pkg/databases/spec"

	"github.com/spf13/cobra"
)

const (
	valueFlag = "value"
	typeFlag  = "type"
)

var (
	allowedTypes = []string{"null", "string", "int32", "int64", "boolean", "uuid"}
)

func encodeCmd() *cobra.Command {
	var value string
	var passedType string

	var cmd = &cobra.Command{
		Use:   "encode",
		Short: "Encode is used to encode a given input type.",
		Long: `The Encode function encodes an input for a specific type.
The type is specified by the --type flag. The value is specified by the --value flag.
The output is printed to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// convert type
			typ, err := spec.DataTypeConversions.StringToKwilType(passedType)
			if err != nil {
				fmt.Printf("type must be one of: %v", allowedTypes)
				return fmt.Errorf("error converting type: %w", err)
			}

			// encode
			val, err := spec.NewExplicit(value, typ)
			if err != nil {
				return fmt.Errorf("error encoding input: %w", err)
			}

			// print
			val.Print()

			return nil
		},
	}

	cmd.Flags().StringVarP(&passedType, typeFlag, "t", "", "The type to encode the value as.")
	cmd.MarkFlagRequired(typeFlag)
	cmd.Flags().StringVarP(&value, valueFlag, "v", "", "The value to encode.")
	cmd.MarkFlagRequired(valueFlag)

	return cmd
}
