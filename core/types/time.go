package types

import "time"

// Duration is a wrapper around time.Duration that implements text
// (un)marshalling for the go-toml package to work with Go duration strings
// instead of integers.
type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}
