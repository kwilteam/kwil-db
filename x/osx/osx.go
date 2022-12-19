package osx

import (
	"kwil/x/utils"
	"os"
	"strings"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

func SetEnv(key string, value string) error {
	return os.Setenv(key, value)
}

func ExpandEnv(s string) string {
	return Expand(s, os.Getenv)
}

func Expand(s string, mapping func(string) string) string {
	const escaped = "$"
	return os.Expand(s, func(key string) string {
		return utils.CoalesceF(utils.IfElse(key == escaped, escaped, ""), func() string {
			parts := strings.SplitN(key, "||", 2)
			value := mapping(strings.TrimSpace(parts[0]))
			if len(parts) == 1 {
				return value
			}

			if value == "" {
				value = strings.TrimSpace(parts[1])
				value = utils.Coalesce(Expand(value, mapping), value)
			}

			_ = os.Setenv(key, value)

			return value
		})
	})
}
