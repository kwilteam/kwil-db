package conv

import (
	"fmt"
	"strconv"
)

func String(a any) (string, error) {
	switch xt := a.(type) {
	case string:
		return xt, nil
	case fmt.Stringer:
		return xt.String(), nil
	case int:
		return signedIntToStr(xt), nil
	case int8:
		return signedIntToStr(xt), nil
	case int16:
		return signedIntToStr(xt), nil
	case int32:
		return signedIntToStr(xt), nil
	case int64:
		return signedIntToStr(xt), nil
	case uint:
		return unsignedIntToStr(xt), nil
	case uint8:
		return unsignedIntToStr(xt), nil
	case uint16:
		return unsignedIntToStr(xt), nil
	case uint32:
		return unsignedIntToStr(xt), nil
	case uint64:
		return unsignedIntToStr(xt), nil
	case uintptr:
		return unsignedIntToStr(xt), nil
	case float32:
		return strconv.FormatFloat(float64(xt), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(xt, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(xt), nil
	}
	return "", fmt.Errorf("cannot convert %T to string", a)
}

func signedIntToStr[T signedInts](val T) string {
	return strconv.FormatInt(int64(val), 10)
}

func unsignedIntToStr[T unsignedInts](val T) string {
	return strconv.FormatUint(uint64(val), 10)
}

type signedInts interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsignedInts interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Int(a any) (int64, error) {
	switch a := a.(type) {
	case int:
		return int64(a), nil
	case int8:
		return int64(a), nil
	case int16:
		return int64(a), nil
	case int32:
		return int64(a), nil
	case int64:
		return a, nil
	case uint:
		return int64(a), nil
	case uint8:
		return int64(a), nil
	case uint16:
		return int64(a), nil
	case uint32:
		return int64(a), nil
	case uint64:
		return int64(a), nil
	case bool:
		if a {
			return 1, nil
		}
		return 0, nil
	case string:
		i, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			return 0, err
		}

		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", a)
	}
}
