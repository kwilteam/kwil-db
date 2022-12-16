package env

import (
	"fmt"
	"os"
	"strings"
)

func GetDbConnectionString() string {
	return GetDbConnectionStringByName("PG_DATABASE_URL", "kwil")
}

func GetDbConnectionStringByName(dbUrlEnvKeyName string, dbNameDefault string) string {
	url := os.Getenv(dbUrlEnvKeyName)
	if url == "" {
		url = fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", dbNameDefault)
	} else {
		url = os.ExpandEnv(url)
	}

	parts := strings.Split(url, "@")
	if len(parts) == 2 {
		fmt.Println("USING DB --> postgres://REDACTED@" + parts[1])
	}

	return url
}
