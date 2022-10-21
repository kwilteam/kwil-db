package cfgx

import "time"

type emptyConfigImpl struct {
}

var emptyConfig = &emptyConfigImpl{}

func (c *emptyConfigImpl) As(_ interface{}) error {
	return nil
}

func (c *emptyConfigImpl) Exists(_ string) bool {
	return false
}

func (c *emptyConfigImpl) ToStringMap() map[string]string {
	return make(map[string]string)
}

func (c *emptyConfigImpl) ToMap() map[string]interface{} {
	return make(map[string]interface{})
}

func (c *emptyConfigImpl) Select(_ string) Config {
	return c
}

func (c *emptyConfigImpl) String(_ string) string {
	return ""
}

func (c *emptyConfigImpl) GetString(_ string, defaultValue string) string {
	return defaultValue
}

func (c *emptyConfigImpl) GetStringSlice(_ string, _ string, defaultValue []string) []string {
	return defaultValue
}

func (c *emptyConfigImpl) StringSlice(_ string, _ string) []string {
	return []string{}
}

func (c *emptyConfigImpl) GetInt32(_ string, defaultValue int32) (int32, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) GetInt64(_ string, defaultValue int64) (int64, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) Int32(_ string, defaultValue int32) int32 {
	return defaultValue
}

func (c *emptyConfigImpl) Int64(_ string, defaultValue int64) int64 {
	return defaultValue
}

func (c *emptyConfigImpl) GetUInt32(_ string, defaultValue uint32) (uint32, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) GetUInt64(_ string, defaultValue uint64) (uint64, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) UInt32(_ string, defaultValue uint32) uint32 {
	return defaultValue
}

func (c *emptyConfigImpl) UInt64(_ string, defaultValue uint64) uint64 {
	return defaultValue
}

func (c *emptyConfigImpl) GetBool(_ string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) Bool(_ string, defaultValue bool) bool {
	return defaultValue
}

func (c *emptyConfigImpl) GetDuration(_ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) Duration(_ string, defaultValue time.Duration) time.Duration {
	return defaultValue
}
