package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBindDefaults(t *testing.T) {
	k = koanf.New(".")

	type PeerConf struct {
		IP       string `custom_tag:"ip" toml:"ip"`
		Port     uint64 `custom_tag:"port" toml:"port"`
		Pex      bool   `custom_tag:"pex" toml:"pex"`
		BootNode string `custom_tag:"bootnode" toml:"bootnode"`
	}

	type testConfig struct {
		LogLevel   string   `custom_tag:"log_level" toml:"log_level"`
		LogFormat  string   `custom_tag:"log_format" toml:"log_format"`
		PrivateKey string   `custom_tag:"privkey" toml:"privkey"`
		PeerConfig PeerConf `custom_tag:"peer" toml:"peer"`
	}

	cfg := &testConfig{
		LogLevel:   "info",
		LogFormat:  "unstructured",
		PrivateKey: "ababababab",
		PeerConfig: PeerConf{
			IP:       "127.0.0.1",
			Port:     6600,
			Pex:      true,
			BootNode: "/ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv",
		},
	}
	if err := BindDefaults(cfg, "custom_tag"); err != nil {
		t.Fatal(err)
	}

	// k.Print()
	assert.Equal(t, "127.0.0.1", k.String("peer.ip"))
	assert.Equal(t, int64(6600), k.Int64("peer.port"))
	assert.Equal(t, true, k.Bool("peer.pex"))
	assert.Equal(t, "/ip4/127.0.0.1/tcp/6600/p2p/16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv", k.String("peer.bootnode"))

}

func TestPreRunBindFlags(t *testing.T) {
	k = koanf.New(".")

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "", "test flag")
	cmd.Flags().Int("number-value", 0, "number value")

	// Set some flag values
	cmd.Flags().Set("test-flag", "test-value")
	cmd.Flags().Set("number-value", "42")

	err := PreRunBindFlags(cmd, []string{})
	assert.NoError(t, err)

	// Verify the values were properly bound (to the underscore converted versions)
	assert.Equal(t, "test-value", k.String("test_flag"))
	assert.Equal(t, 42, k.Int("number_value"))
}

func TestPreRunBindEnvMatching(t *testing.T) {
	k = koanf.New(".")
	// k.Set("top", "default-top")
	k.Set("test.value", "default-value")
	k.Set("nested.section.value", "default-nested-value")
	k.Set("nested.section.long-value", "default-nested-long-value")

	// Set test environment variables
	os.Setenv("KWIL_TEST_VALUE", "env-test")
	os.Setenv("KWIL_NESTED_SECTION_VALUE", "nested-value")
	os.Setenv("KWIL_NESTED_SECTION_LONG_VALUE", "nested-long-value")
	defer os.Unsetenv("KWIL_TEST_VALUE")
	defer os.Unsetenv("KWIL_NESTED_SECTION_VALUE")
	defer os.Unsetenv("KWIL_NESTED_SECTION_LONG_VALUE")

	cmd := &cobra.Command{Use: "test"}
	err := PreRunBindEnvMatching(cmd, []string{})
	assert.NoError(t, err)

	// k.Print()
	// nested.section.long-value -> nested-long-value
	// nested.section.value -> nested-value
	// test.value -> env-test

	// Verify environment variables were properly bound
	assert.Equal(t, "env-test", k.String("test.value"))
	assert.Equal(t, "nested-value", k.String("nested.section.value"))
}

func TestPreRunBindConfigFile(t *testing.T) {
	k = koanf.New(".")

	tmpDir := t.TempDir()

	configContent := `
top_val = "a"

[section]
key = "value"
number = 42

[nested]
string_value = "nested-string"
`
	configPath := filepath.Join(tmpDir, ConfigFileName)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Create test command with root flag
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("root", tmpDir, "root directory")

	err = PreRunBindConfigFile(cmd, []string{})
	assert.NoError(t, err)

	// Verify config values were properly loaded
	assert.Equal(t, "a", k.String("top_val"))
	assert.Equal(t, "value", k.String("section.key"))
	assert.Equal(t, 42, k.Int("section.number"))
	assert.Equal(t, "nested-string", k.String("nested.string_value"))
}

func TestPreRunBindConfigFileNonExistent(t *testing.T) {
	k = koanf.New(".")

	// Create temporary directory without config file
	tmpDir, err := os.MkdirTemp("", "kwil-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("root", tmpDir, "root directory")

	// Should not error when config file doesn't exist
	err = PreRunBindConfigFile(cmd, []string{})
	assert.NoError(t, err)
}

func TestMergeFunc(t *testing.T) {
	t.Run("merge simple maps", func(t *testing.T) {
		src := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		dest := map[string]interface{}{
			"key3": "value3",
		}

		err := mergeFunc(src, dest, func(s string) string { return s })
		assert.NoError(t, err)
		assert.Equal(t, "value1", dest["key1"])
		assert.Equal(t, 42, dest["key2"])
		assert.Equal(t, "value3", dest["key3"])
	})

	t.Run("merge nested maps", func(t *testing.T) {
		src := map[string]interface{}{
			"nested": map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		}
		dest := map[string]interface{}{
			"nested": map[string]interface{}{
				"c": 3,
			},
		}

		err := mergeFunc(src, dest, func(s string) string { return s })
		assert.NoError(t, err)

		expected := map[string]interface{}{
			"nested": map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		}
		assert.Equal(t, expected, dest)
	})

	t.Run("key transformation", func(t *testing.T) {
		src := map[string]interface{}{
			"key-one": "value1",
			"key-two": "value2",
		}
		dest := map[string]interface{}{}

		err := mergeFunc(src, dest, func(s string) string {
			return strings.ReplaceAll(s, "-", "_")
		})
		assert.NoError(t, err)
		assert.Equal(t, "value1", dest["key_one"])
		assert.Equal(t, "value2", dest["key_two"])
	})

	t.Run("conflict error", func(t *testing.T) {
		src := map[string]interface{}{
			"key": map[string]interface{}{
				"nested": "value",
			},
		}
		dest := map[string]interface{}{
			"key": "simple_value",
		}

		err := mergeFunc(src, dest, func(s string) string { return s })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflict")
	})

	t.Run("empty source map", func(t *testing.T) {
		src := map[string]interface{}{}
		dest := map[string]interface{}{
			"existing": "value",
		}

		err := mergeFunc(src, dest, func(s string) string { return s })
		assert.NoError(t, err)
		assert.Equal(t, "value", dest["existing"])
	})

	t.Run("overwrite existing values", func(t *testing.T) {
		src := map[string]interface{}{
			"key": "new_value",
		}
		dest := map[string]interface{}{
			"key": "old_value",
		}

		err := mergeFunc(src, dest, func(s string) string { return s })
		assert.NoError(t, err)
		assert.Equal(t, "new_value", dest["key"])
	})
}

func TestSetNodeFlagsFromStruct(t *testing.T) {
	type NestedConfig struct {
		Host     string  `toml:"host" comment:"Host address"`
		Port     uint16  `toml:"port" comment:"Port number"`
		Enabled  bool    `toml:"enabled"`
		Ratio    float64 `toml:"ratio"`
		Tags     []string
		Priority int32 `toml:"priority"`
	}

	type TestConfig struct {
		Name            string       `toml:"name" comment:"Service name"`
		Version         int          `toml:"version"`
		Debug           bool         `toml:"debug" comment:"Enable debug mode"`
		Nested          NestedConfig `toml:"nested"`
		UnderscoredName string       `toml:"underscored_name" comment:"Custom service name"`
		UntaggedName    string       `comment:"untagged service name"`
	}

	t.Run("basic flag creation", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cfg := TestConfig{}
		SetNodeFlagsFromStruct(cmd, cfg)

		flags := cmd.Flags()
		assert.NotNil(t, flags.Lookup("name"))
		assert.NotNil(t, flags.Lookup("version"))
		assert.NotNil(t, flags.Lookup("debug"))
		assert.NotNil(t, flags.Lookup("underscored-name"))
		assert.NotNil(t, flags.Lookup("untaggedname"))
	})

	t.Run("nested struct flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cfg := TestConfig{}
		SetNodeFlagsFromStruct(cmd, cfg)

		flags := cmd.Flags()
		assert.NotNil(t, flags.Lookup("nested.host"))
		assert.NotNil(t, flags.Lookup("nested.port"))
		assert.NotNil(t, flags.Lookup("nested.enabled"))
		assert.NotNil(t, flags.Lookup("nested.ratio"))
		assert.NotNil(t, flags.Lookup("nested.tags"))
		assert.NotNil(t, flags.Lookup("nested.priority"))
	})

	t.Run("flag descriptions", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cfg := TestConfig{}
		SetNodeFlagsFromStruct(cmd, cfg)

		flags := cmd.Flags()
		assert.Equal(t, "Service name", flags.Lookup("name").Usage)
		assert.Equal(t, "Enable debug mode", flags.Lookup("debug").Usage)
		assert.Equal(t, "Host address", flags.Lookup("nested.host").Usage)
		assert.Equal(t, "nested.ratio", flags.Lookup("nested.ratio").Usage)
	})

	t.Run("flag types", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cfg := TestConfig{}
		SetNodeFlagsFromStruct(cmd, cfg)

		flags := cmd.Flags()
		assert.Equal(t, "string", flags.Lookup("name").Value.Type())
		assert.Equal(t, "int64", flags.Lookup("version").Value.Type())
		assert.Equal(t, "bool", flags.Lookup("debug").Value.Type())
		assert.Equal(t, "uint64", flags.Lookup("nested.port").Value.Type())
		assert.Equal(t, "float64", flags.Lookup("nested.ratio").Value.Type())
		assert.Equal(t, "stringSlice", flags.Lookup("nested.tags").Value.Type())
	})

	t.Run("empty struct", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		type EmptyConfig struct{}
		cfg := EmptyConfig{}
		SetNodeFlagsFromStruct(cmd, cfg)
		assert.Zero(t, cmd.Flags().NFlag())
	})

	t.Run("default values", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		type DefaultConfig struct {
			Host     string  `toml:"host" comment:"Host address"`
			Port     uint16  `toml:"port" comment:"Port number"`
			Enabled  bool    `toml:"enabled"`
			Ratio    float64 `toml:"ratio"`
			Tags     []string
			Priority int32 `toml:"priority"`
		}
		cfg := DefaultConfig{
			Host:     "localhost",
			Port:     8080,
			Enabled:  true,
			Ratio:    0.5,
			Tags:     []string{"tag1", "tag2"},
			Priority: 10,
		}
		SetNodeFlagsFromStruct(cmd, cfg)

		flags := cmd.Flags()
		assert.Equal(t, "localhost", flags.Lookup("host").DefValue)
		assert.Equal(t, "8080", flags.Lookup("port").DefValue)
		assert.Equal(t, "true", flags.Lookup("enabled").DefValue)
		assert.Equal(t, "0.5", flags.Lookup("ratio").DefValue)
		assert.Equal(t, "[tag1,tag2]", flags.Lookup("tags").DefValue)
		assert.Equal(t, "10", flags.Lookup("priority").DefValue)
	})
}
