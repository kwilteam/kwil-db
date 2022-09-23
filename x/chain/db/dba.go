package db

import (
	"path/filepath"
	"strings"

	ms "github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	types "kwil/pkg/types/db"
)

func LoadSQLConfig(path string) (*types.SqlDatabaseConfig, error) {
	var dbConfig types.SqlDatabaseConfig
	logger := log.With().Str("module", "dba").Logger()

	dir, file := filepath.Split(path)
	strs := strings.Split(file, ".")

	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	viper.SetConfigType(strs[len(strs)-1])

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read config")
	}

	err = viper.Unmarshal(&dbConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to unmarshal config")
	}

	// Since there are several different constraints, they come as map[string]interface{}
	// We must convert them to the correct type
	for _, constraint := range dbConfig.Structure.MappedConstraints {
		switch constraint["type"] {
		case "primary_key":
			// Currently not supported, but needs to be added ASAP
		case "foreign_key":
			fkConst, err := UnloadForeignKey(constraint)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to unmarshal foreign key constraint")
			}
			dbConfig.Structure.Constraints = append(dbConfig.Structure.Constraints, fkConst)
		}
	}
	return &dbConfig, nil
}

func UnloadForeignKey(c map[string]interface{}) (types.ForeignKeyConstraint, error) {
	var fk types.ForeignKeyConstraint

	// ms package is used to unmarshal a map to a struct
	err := ms.Decode(c, &fk)
	if err != nil {
		return fk, err
	}
	return fk, nil
}
