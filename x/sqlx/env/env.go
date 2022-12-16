package env

import (
	"fmt"
	"kwil/x/utils"
	"os"
	"strings"
)

func GetDbConnectionString() string {
	return GetAltDbConnectionString(os.Getenv("PG_DB"), "kwil")
}

func GetAltDbConnectionString(dbEnvKey string, defaultDbName string) string {
	host := utils.Coalesce(os.Getenv("PG_ENDPOINT"), "localhost")
	port := utils.Coalesce(os.Getenv("PG_PORT"), "5432")
	user := utils.Coalesce(os.Getenv("PG_USER"), "postgres")
	password := utils.Coalesce(os.Getenv("PG_PASSWORD"), "postgres")
	database := utils.Coalesce(os.Getenv(dbEnvKey), defaultDbName)

	var ssl string
	if strings.HasPrefix(host, "postgres_db_container_local") || strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		ssl = "disable"
	} else {
		ssl = "require"
	}

	url := fmt.Sprintf("USING DB --> postgres://%s:%s@%s:%s/%s?sslmode=%s", user, "REDACTED", host, port, database, ssl)
	fmt.Println(url)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, database, ssl)
}
