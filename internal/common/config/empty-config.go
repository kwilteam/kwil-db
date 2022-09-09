package config

type emptyConfigImpl struct {
}

var emptyConfig = &emptyConfigImpl{}

func (c *emptyConfigImpl) As(out interface{}) error {
	return nil
}

func (c *emptyConfigImpl) Exists(key string) bool {
	return false
}

func (c *emptyConfigImpl) ToMap() map[string]string {
	return make(map[string]string)
}

func (c *emptyConfigImpl) Select(prefix string) Config {
	return c
}

func (c *emptyConfigImpl) String(key string) string {
	return ""
}

func (c *emptyConfigImpl) GetString(key string, defaultValue string) string {
	return defaultValue
}

func (c *emptyConfigImpl) GetInt32(key string, defaultValue int32) (int32, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) GetInt64(key string, defaultValue int64) (int64, error) {
	return defaultValue, nil
}

func (c *emptyConfigImpl) Int32(key string, defaultValue int32) int32 {
	return defaultValue
}

func (c *emptyConfigImpl) Int64(key string, defaultValue int64) int64 {
	return defaultValue
}
