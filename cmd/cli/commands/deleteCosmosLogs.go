package commands

import (
	"os"
)

// Function deletes all logs in /kwil-cosmos-logger/tmp and /kwil-cosmos-logger/finished_logs
func DeleteLogs() error {
	err := os.RemoveAll("./kwil-cosmos-logger/tmp")
	if err != nil {
		return err
	}
	err = os.MkdirAll("./kwil-cosmos-logger/tmp", 0755)
	if err != nil {
		return err
	}
	err = os.RemoveAll("./kwil-cosmos-logger/finished_logs")
	if err != nil {
		return err
	}
	return os.MkdirAll("./kwil-cosmos-logger/finished_logs", 0755)
}
