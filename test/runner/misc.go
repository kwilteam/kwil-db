package runner

import "os"

func GetEnv(key, defalutValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defalutValue
	}
	return v
}
