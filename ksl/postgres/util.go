package postgres

import (
	"fmt"
	"ksl/sqlmigrate"
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

func columnTypeChange(prev, next sqlschema.ColumnWalker) sqlmigrate.ColumnTypeChange {
	prevEnum, prevIsEnum := prev.Type().Type.(sqlschema.EnumType)
	nextEnum, nextIsEnum := next.Type().Type.(sqlschema.EnumType)
	switch {
	case prevIsEnum && nextIsEnum && prevEnum.ID == nextEnum.ID:
		return sqlmigrate.ColumnTypeChangeNone
	case prevIsEnum && nextIsEnum && prevEnum.ID != nextEnum.ID:
		return sqlmigrate.ColumnTypeChangeNotCastable
	case prevIsEnum != nextIsEnum:
		return sqlmigrate.ColumnTypeChangeNotCastable
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
			return sqlmigrate.ColumnTypeChangeSafeCast
		case name == TypeVarChar && len(args) == 0 && fromListToScalar:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case name == TypeVarChar && len(args) > 0 && fromListToScalar:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case name == TypeChar && len(args) > 0 && fromListToScalar:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case fromScalarToList || fromListToScalar:
			return sqlmigrate.ColumnTypeChangeNotCastable
		case prev.Type().Raw == next.Type().Raw:
			return sqlmigrate.ColumnTypeChangeNone
		default:
			if prevType, ok := prevType.(PostgresType); ok {
				return columnTypeChangeRiskiness(prevType, typ)
			}
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}
	default:
		if prev.Type().Raw == next.Type().Raw {
			return sqlmigrate.ColumnTypeChangeNone
		}
		return sqlmigrate.ColumnTypeChangeRiskyCast
	}
}

func columnTypeChangeRiskiness(prev, next PostgresType) sqlmigrate.ColumnTypeChange {
	if prev.String() == next.String() {
		return sqlmigrate.ColumnTypeChangeNone
	}

	switch left, right := AliasType(prev.Name()), AliasType(next.Name()); left {
	case TypeInet:
		switch right {
		case TypeCIText, TypeText, TypeVarChar:
			return sqlmigrate.ColumnTypeChangeSafeCast
		default:
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

	case TypeMoney:
		switch right {
		case TypeNumeric:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeText, TypeCIText, TypeVarChar:
			return sqlmigrate.ColumnTypeChangeSafeCast
		default:
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

	case TypeCIText:
		switch right {
		case TypeText, TypeVarChar:
			return sqlmigrate.ColumnTypeChangeSafeCast
		default:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}

	case TypeSmallInt:
		switch right {
		case TypeInteger, TypeBigInt, TypeReal, TypeDouble, TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 3:
				// SmallInt can be at most three digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 3:
				// SmallInt can be at most three digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 4 {
				// Smallint can have three digits and an optional sign.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 4) {
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

	case TypeInteger:
		switch right {
		case TypeSmallInt:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeBigInt, TypeReal, TypeDouble, TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 10:
				// Integer can be at most 10 digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 10:
				// Integer can be at most 10 digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 11 {
				// Integer can have 10 digits and an optional sign.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 11) {
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

	case TypeBigInt:
		switch right {
		case TypeSmallInt, TypeInteger:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeReal, TypeDouble, TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeNumeric:
			args := stoi(next.Args())
			switch {
			case len(args) == 2 && args[0]-args[1] < 19:
				// Bigint can be at most nineteen digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(args) == 1 && args[0] < 19:
				// Bigint can be at most nineteen digits, so this might fail.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar:
			if args := stoi(next.Args()); len(args) == 1 && args[0] < 20 {
				// Integer can have 19 digits and an optional sign.
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeChar:
			if args := stoi(next.Args()); len(args) == 0 || (len(args) == 1 && args[0] < 20) {
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		default:
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

	case TypeNumeric:
		params := stoi(prev.Args())
		switch right {
		case TypeSmallInt:
			switch {
			case len(params) > 0 && params[0] > 2:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeInteger:
			switch {
			case len(params) > 0 && params[0] > 9:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeBigInt:
			switch {
			case len(params) > 0 && params[0] > 18:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) > 1 && params[1] > 0:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeNumeric:
			np := stoi(next.Args())
			switch {
			case len(params) > 0 && len(np) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 0 && len(np) == 2 && (np[0] < 131072 || np[1] < 16383):
				return sqlmigrate.ColumnTypeChangeRiskyCast
			// So, numeric(4,0) to numeric(4,2) would be risky, so would numeric(4,2) to numeric(4,0).
			case len(params) == 2 && len(np) == 2 && ((params[0]-params[1] > np[0]-np[1]) || params[1] > np[1]):
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast

		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			// We must fit p digits and a possible sign to our string, otherwise might truncate.
			case len(params) == 1 && len(np) == 1 && np[0]+1 > params[0]:
				return sqlmigrate.ColumnTypeChangeRiskyCast

			// We must fit p digits, a possible sign and a comma to our string, otherwise might truncate.
			case len(params) == 1 && len(np) == 2 && np[1] > 0 && np[0]+2 > params[0]:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			// up to 131072 digits before the decimal point; up to 16383 digits after the decimal point
			case len(params) == 1 && len(np) == 0 && params[0] < 131073:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeReal, TypeDouble:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeReal:
		switch right {
		case TypeSmallInt, TypeInteger, TypeBigInt, TypeNumeric:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeReal, TypeDouble, TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] < 47:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(np) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable
	case TypeDouble:
		switch right {
		case TypeSmallInt, TypeInteger, TypeBigInt, TypeNumeric, TypeReal:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeDouble, TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] < 317:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(np) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable
	case TypeVarChar:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(params) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) == 0 && len(np) == 1:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) == 1 && params[0] == 1 && len(np) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) == 1 && len(np) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 1 && params[0] > np[0]:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeChar:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(params) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 1 && params[0] == 1 && len(np) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 0 && right == TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(params) == 1 && len(np) == 0:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(params) == 1 && len(np) == 1 && params[0] > np[0]:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeText:
		switch right {
		case TypeText, TypeCIText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeChar:
			return sqlmigrate.ColumnTypeChangeRiskyCast
		case TypeVarChar:
			if len(next.Args()) == 0 {
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeByteA:
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 2:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeTimestamp:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 22:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		case TypeTimestamp:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlmigrate.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlmigrate.ColumnTypeChangeNone
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeTimestampTZ, TypeDate, TypeTime, TypeTimeTZ:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeTimestampTZ:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 27:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		case TypeTimestampTZ:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlmigrate.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlmigrate.ColumnTypeChangeNone
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeTimestamp, TypeDate, TypeTime, TypeTimeTZ:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeDate:
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 27:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		case TypeTimestamp, TypeTimestampTZ:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeTime:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 13:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		case TypeTime:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlmigrate.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlmigrate.ColumnTypeChangeNone
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeTimeTZ:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeTimeTZ:
		params := stoi(prev.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 18:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		case TypeTimeTZ:
			switch np := stoi(prev.Args()); {
			case len(np) == 0:
				return sqlmigrate.ColumnTypeChangeNone
			case len(params) == 0 || params[0] == np[0]:
				return sqlmigrate.ColumnTypeChangeNone
			}
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeTime:
			return sqlmigrate.ColumnTypeChangeSafeCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeBoolean:
		switch right {
		case TypeText, TypeVarChar:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeChar:
			np := stoi(next.Args())
			switch {
			case len(np) == 1 && np[0] > 4:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 3:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeBit:
		params := stoi(prev.Args())
		if len(params) == 0 {
			switch right {
			case TypeText, TypeVarChar, TypeChar, TypeVarBit:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

		np := stoi(next.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeVarBit, TypeChar:
			switch {
			case len(np) == 0 && right != TypeChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] >= params[0]:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeVarBit:
		params := stoi(prev.Args())
		np := stoi(next.Args())

		if len(params) == 0 {
			switch right {
			case TypeText:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case TypeVarChar, TypeChar, TypeBit:
				if right == TypeVarChar && len(np) == 0 {
					return sqlmigrate.ColumnTypeChangeSafeCast
				}
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
			return sqlmigrate.ColumnTypeChangeNotCastable
		}

		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeVarBit, TypeChar:
			switch {
			case len(np) == 0 && right != TypeChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && right == TypeVarBit && np[0] > params[0]:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && right != TypeVarBit && np[0] >= params[0]:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			switch right {
			case TypeVarChar, TypeChar:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		case TypeBit:
			switch {
			case len(np) == 0:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			case len(np) == 1 && np[0] <= params[0]:
				return sqlmigrate.ColumnTypeChangeRiskyCast
			}
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeUUID:
		np := stoi(next.Args())
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(np) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			case len(np) == 1 && np[0] > 31:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeXML:
		switch right {
		case TypeText:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeJson:
		switch right {
		case TypeText, TypeJsonB:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable

	case TypeJsonB:
		switch right {
		case TypeText, TypeJson:
			return sqlmigrate.ColumnTypeChangeSafeCast
		case TypeVarChar, TypeChar:
			switch {
			case len(next.Args()) == 0 && right == TypeVarChar:
				return sqlmigrate.ColumnTypeChangeSafeCast
			}
			return sqlmigrate.ColumnTypeChangeRiskyCast
		}
		return sqlmigrate.ColumnTypeChangeNotCastable
	}

	return sqlmigrate.ColumnTypeChangeNotCastable
}
