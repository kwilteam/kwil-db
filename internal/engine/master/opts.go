package master

import (
	"fmt"
	"os"
)

const (
	defaultName = "kwil_master"
)

var defaultPath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

type MasterOpt func(*MasterDB)

// WithPath sets the path option.
func WithPath(path string) MasterOpt {
	return func(m *MasterDB) {
		m.path = path
	}
}

// WithFileName sets the name option.
func WithFileName(name string) MasterOpt {
	return func(m *MasterDB) {
		m.name = name
	}
}

type DbidFunc func(name string, owner []byte) string

// WithDbidFunc sets the DbidFunc option.
func WithDbidFunc(f DbidFunc) MasterOpt {
	return func(m *MasterDB) {
		m.DbidFunc = f
	}
}
