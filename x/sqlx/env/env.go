package env

import (
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/osx"
	"kwil/x/utils"
	"strings"
)

func GetDbConnectionString() string {
	return GetDbConnectionStringByName("PG_DATABASE_URL", "kwil")
}

func GetDbConnectionStringByName(dbUrlEnvKeyName string, dbNameDefault string) string {
	url := osx.GetEnv(dbUrlEnvKeyName)
	if url == "" {
		url = utils.Coalesce(getDbStringFromMetaConfig(), "postgres://postgres:postgres@localhost:5432/%s?sslmode=disable")
	} else {
		url = osx.ExpandEnv(url)
	}

	if strings.Index("%s", url) != -1 {
		url = fmt.Sprintf(url, dbNameDefault)
	}

	parts := strings.Split(url, "@")
	if len(parts) == 2 {
		fmt.Println("USING DB --> postgres://REDACTED@" + parts[1])
	}

	return url
}

func getDbStringFromMetaConfig() string {
	cfg := cfgx.GetConfig().Select("db-settings")

	host := cfg.String("host")
	if host == "" {
		return ""
	}

	port := cfg.Int32("port", 5432)
	user := cfg.String("user")
	password := cfg.String("password")
	database := cfg.String("database")
	ssl_mode := cfg.GetString("ssl_mode", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, password, host, port, database, ssl_mode)
}
