package postgres

import (
	"fmt"
	"ksl/sqlschema"
	"strconv"
)

func itos(args []int) []string {
	var result []string
	for _, arg := range args {
		result = append(result, strconv.Itoa(arg))
	}
	return result
}

func stoi(args []string) []int {
	result := make([]int, 0, len(args))
	for _, arg := range args {
		val, err := strconv.Atoi(arg)
		if err != nil {
			continue
		}
		result = append(result, val)
	}
	return result
}

func quoteIdent(s ...string) string {
	var ss string
	for i, v := range s {
		if i > 0 {
			ss += ", "
		}
		ss += fmt.Sprintf(`"%s"`, v)
	}
	return ss
}

func quoteString(s ...string) string {
	var ss string
	for i, v := range s {
		if i > 0 {
			ss += ", "
		}
		ss += fmt.Sprintf(`'%s'`, v)
	}
	return ss
}

func columnTypeChange(prev, next sqlschema.ColumnWalker) sqlschema.ColumnTypeChange {
	prevEnum, prevIsEnum := prev.Type().Type.(sqlschema.EnumType)
	nextEnum, nextIsEnum := next.Type().Type.(sqlschema.EnumType)
	switch {
	case prevIsEnum && nextIsEnum && prevEnum.ID == nextEnum.ID:
		return sqlschema.ColumnTypeChangeNone
	case prevIsEnum && nextIsEnum && prevEnum.ID != nextEnum.ID:
		return sqlschema.ColumnTypeChangeNotCastable
	case prevIsEnum != nextIsEnum:
		return sqlschema.ColumnTypeChangeNotCastable
	}

	prevType, nextType := prev.Type().Type, next.Type().Type
	fromListToScalar := prev.Arity() == sqlschema.List && next.Arity() != sqlschema.List
	fromScalarToList := prev.Arity() != sqlschema.List && next.Arity() == sqlschema.List

	switch typ := nextType.(type) {
	case PostgresType:
		name := typ.Name()
		args := typ.Args()

		switch {
		case name == TypeText && fromListToScalar:
			return sqlschema.ColumnTypeChangeSafeCast
		case name == TypeVarChar && len(args) == 0 && fromListToScalar:
			return sqlschema.ColumnTypeChangeSafeCast
		case name == TypeVarChar && len(args) > 0 && fromListToScalar:
			return sqlschema.ColumnTypeChangeRiskyCast
		case name == TypeChar && len(args) > 0 && fromListToScalar:
			return sqlschema.ColumnTypeChangeRiskyCast
		case fromScalarToList || fromListToScalar:
			return sqlschema.ColumnTypeChangeNotCastable
		case prev.Type().Raw == next.Type().Raw:
			return sqlschema.ColumnTypeChangeNone
		default:
			if prevType, ok := prevType.(PostgresType); ok {
				return columnTypeChangeRiskiness(prevType, typ)
			}
			return sqlschema.ColumnTypeChangeRiskyCast
		}
	default:
		if prev.Type().Raw == next.Type().Raw {
			return sqlschema.ColumnTypeChangeNone
		}
		return sqlschema.ColumnTypeChangeRiskyCast
	}
}

func columnTypeChangeRiskiness(prev, next PostgresType) sqlschema.ColumnTypeChange {
	if prev.String() == next.String() {
		return sqlschema.ColumnTypeChangeNone
	}

	switch left, right := AliasType(prev.Name()), AliasType(next.Name()); left {
	case TypeInet:
		switch right {
		case TypeCIText, TypeText, TypeVarChar:
			return sqlschema.ColumnTypeChangeSafeCast
		default:
			return sqlschema.ColumnTypeChangeNotCastable
		}

	case TypeMoney:
		switch right {
		case TypeNumeric:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeText, TypeCIText, TypeVarChar:
			return sqlschema.ColumnTypeChangeSafeCast
		default:
			return sqlschema.ColumnTypeChangeNotCastable
		}

	case TypeCIText:
		switch right {
		case TypeText, TypeVarChar:
			return sqlschema.ColumnTypeChangeSafeCast
		default:
			return sqlschema.ColumnTypeChangeRiskyCast
		}

	case TypeSmallInt:
		switch right {
		case TypeInteger, TypeBigInt, TypeReal, TypeDouble, TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 3:
				// SmallInt can be at most three digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 3:
				// SmallInt can be at most three digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 4 {
				// Smallint can have three digits and an optional sign.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 4) {
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlschema.ColumnTypeChangeNotCastable
		}

	case TypeInteger:
		switch right {
		case TypeSmallInt:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeBigInt, TypeReal, TypeDouble, TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 10:
				// Integer can be at most 10 digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 10:
				// Integer can be at most 10 digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 11 {
				// Integer can have 10 digits and an optional sign.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 11) {
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlschema.ColumnTypeChangeNotCastable
		}

	case TypeBigInt:
		switch right {
		case TypeSmallInt, TypeInteger:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeReal, TypeDouble, TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 19:
				// Bigint can be at most nineteen digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 19:
				// Bigint can be at most nineteen digits, so this might fail.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 20 {
				// Integer can have 19 digits and an optional sign.
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 20) {
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlschema.ColumnTypeChangeNotCastable
		}

	case TypeNumeric:
		params := stoi(prev.Args())
		switch right {
		case TypeSmallInt:
			switch {
			case len(params) > 0 && params[0] > 2:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeInteger:
			switch {
			case len(params) > 0 && params[0] > 9:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeBigInt:
			switch {
			case len(params) > 0 && params[0] > 18:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeNumeric:
			np := stoi(next.Args())
			switch {
			case len(params) > 0 && len(np) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 0 && len(np) == 2 && (np[0] < 131072 || np[1] < 16383):
				return sqlschema.ColumnTypeChangeRiskyCast
			// So, numeric(4,0) to numeric(4,2) would be risky, so would numeric(4,2) to numeric(4,0).
			case len(params) == 2 && len(np) == 2 && ((params[0]-params[1] > np[0]-np[1]) || params[1] > np[1]):
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast

		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			// We must fit p digits and a possible sign to our string, otherwise might truncate.
			case len(params) == 1 && len(np) == 1 && np[0]+1 > params[0]:
				return sqlschema.ColumnTypeChangeRiskyCast

			// We must fit p digits, a possible sign and a comma to our string, otherwise might truncate.
			case len(params) == 1 && len(np) == 2 && np[1] > 0 && np[0]+2 > params[0]:
				return sqlschema.ColumnTypeChangeRiskyCast
			// up to 131072 digits before the decimal point; up to 16383 digits after the decimal point
			case len(params) == 1 && len(np) == 0 && params[0] < 131073:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeReal, TypeDouble:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeReal:
		switch right {
		case TypeSmallInt, TypeInteger, TypeBigInt, TypeNumeric:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeReal, TypeDouble, TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] < 47:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(np) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable
	case TypeDouble:
		switch right {
		case TypeSmallInt, TypeInteger, TypeBigInt, TypeNumeric, TypeReal:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeDouble, TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] < 317:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(np) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable
	case TypeVarChar:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(params) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) == 0 && len(np) == 1:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) == 1 && params[0] == 1 && len(np) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) == 1 && len(np) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 1 && params[0] > np[0]:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeChar:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(params) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 1 && params[0] == 1 && len(np) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 0 && right == TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(params) == 1 && len(np) == 0:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 1 && params[0] > np[0]:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeText:
		switch right {
		case TypeText, TypeCIText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeChar:
			return sqlschema.ColumnTypeChangeRiskyCast
		case TypeVarChar:
			if len(next.Args()) == 0 {
				return sqlschema.ColumnTypeChangeSafeCast
			}
			return sqlschema.ColumnTypeChangeRiskyCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeByteA:
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 2:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeTimestamp:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 22:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		case TypeTimestamp:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlschema.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlschema.ColumnTypeChangeNone
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeTimestampTZ, TypeDate, TypeTime, TypeTimeTZ:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeTimestampTZ:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 27:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		case TypeTimestampTZ:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlschema.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlschema.ColumnTypeChangeNone
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeTimestamp, TypeDate, TypeTime, TypeTimeTZ:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeDate:
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 27:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		case TypeTimestamp, TypeTimestampTZ:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeTime:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 13:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		case TypeTime:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlschema.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlschema.ColumnTypeChangeNone
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeTimeTZ:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeTimeTZ:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 18:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		case TypeTimeTZ:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlschema.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlschema.ColumnTypeChangeNone
			}
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeTime:
			return sqlschema.ColumnTypeChangeSafeCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeBoolean:
		switch right {
		case TypeText, TypeVarChar:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] > 4:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 3:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeBit:
		params := stoi(prev.Args())
		if len(params) == 0 {
			switch right {
			case TypeText, TypeVarChar, TypeChar, TypeVarBit:
				return sqlschema.ColumnTypeChangeSafeCast
			}
			return sqlschema.ColumnTypeChangeNotCastable
		}

		np := stoi(next.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeVarBit, TypeChar:
			switch {
			case len(np) == 0 && right != TypeChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] >= params[0]:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeVarBit:
		params := stoi(prev.Args())
		np := stoi(next.Args())

		if len(params) == 0 {
			switch right {
			case TypeText:
				return sqlschema.ColumnTypeChangeSafeCast
			case TypeVarChar, TypeChar, TypeBit:
				if right == TypeVarChar && len(np) == 0 {
					return sqlschema.ColumnTypeChangeSafeCast
				}
				return sqlschema.ColumnTypeChangeRiskyCast
			}
			return sqlschema.ColumnTypeChangeNotCastable
		}

		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeVarBit, TypeChar:
			switch {
			case len(np) == 0 && right != TypeChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && right == TypeVarBit && np[0] > params[0]:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && right != TypeVarBit && np[0] >= params[0]:
				return sqlschema.ColumnTypeChangeSafeCast
			}
			switch right {
			case TypeVarChar, TypeChar:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		case TypeBit:
			switch {
			case len(np) == 0:
				return sqlschema.ColumnTypeChangeRiskyCast
			case len(np) == 1 && np[0] <= params[0]:
				return sqlschema.ColumnTypeChangeRiskyCast
			}
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeUUID:
		np := stoi(next.Args())
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 31:
				return sqlschema.ColumnTypeChangeSafeCast
			}
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeXML:
		switch right {
		case TypeText:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			}
			return sqlschema.ColumnTypeChangeRiskyCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeJson:
		switch right {
		case TypeText, TypeJsonB:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			}
			return sqlschema.ColumnTypeChangeRiskyCast
		}
		return sqlschema.ColumnTypeChangeNotCastable

	case TypeJsonB:
		switch right {
		case TypeText, TypeJson:
			return sqlschema.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlschema.ColumnTypeChangeSafeCast
			}
			return sqlschema.ColumnTypeChangeRiskyCast
		}
		return sqlschema.ColumnTypeChangeNotCastable
	}

	return sqlschema.ColumnTypeChangeNotCastable
}
