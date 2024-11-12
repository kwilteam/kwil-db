package utils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"kwil/utils"
)

// testing dbid is case insensitive
func Test_DBID(t *testing.T) {
	owner := []byte("owner")
	name := "name"

	dbid1 := utils.GenerateDBID(name, owner)
	dbid2 := utils.GenerateDBID(strings.ToUpper(name), owner)

	assert.Equal(t, dbid1, dbid2)
}
