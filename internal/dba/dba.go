package dba

import (
	types "github.com/kwilteam/kwil-db/pkg/types/dba"
	ms "github.com/mitchellh/mapstructure"
	//	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

func LoadSQLConfig(path string) (*types.SqlDatabaseConfig, error) {
	var dbConfig types.SqlDatabaseConfig

	dir, file := filepath.Split(path)
	strs := strings.Split(file, ".")

	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	viper.SetConfigType(strs[len(strs)-1])

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err // Returning empty config if error occurs
	}

	err = viper.Unmarshal(&dbConfig)
	if err != nil {
		return nil, err
	}

	// Since there are several different constraints, they come as map[string]interface{}
	// We must convert them to the correct type
	for _, constraint := range dbConfig.Structure.MappedConstraints {
		switch constraint["type"] {
		case "primary_key":
			// Currently not supported, but needs to be added ASAP
		case "foreign_key":
			fkConst := UnloadForeignKey(constraint)
			dbConfig.Structure.Constraints = append(dbConfig.Structure.Constraints, fkConst)
		}
	}
	return &dbConfig, nil
}

func UnloadForeignKey(c map[string]interface{}) types.ForeignKeyConstraint {
	var fk types.ForeignKeyConstraint
	ms.Decode(c, &fk)
	return fk
}
