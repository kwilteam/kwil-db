package mocks

import (
	"github.com/kwilteam/kwil-db/pkg/databases/executables"
)

var (
	// it is important this corresponds to the queries in kwil.go

	// Insert1 and Insert2 (have 1 non-static)
	Insert1Inputs = []*executables.UserInput{&Param2Input}
	Insert2Inputs = []*executables.UserInput{&Param3Input}

	// Update1 (has 1 non-static)
	Update1Inputs = []*executables.UserInput{&Param2Input}

	// Update2 (has 2 non-static)
	Update2Inputs = []*executables.UserInput{&Param3Input, &Where1Input}

	// Delete1 (has 0 non-static)
	Delete1Inputs = []*executables.UserInput{}

	// Delete2 (has 1 non-static)
	Delete2Inputs = []*executables.UserInput{&Where1Input}

	Param2Input = executables.UserInput{
		Name:  Parameter2.Name,
		Value: []byte{3, 164, 1, 0, 0},
	}

	// Insert2 (has 1 non-static)
	Param3Input = executables.UserInput{
		Name:  Parameter3.Name,
		Value: []byte{5, 1},
	}

	Where1Input = executables.UserInput{
		Name:  WhereClause1.Name,
		Value: []byte{5, 1},
	}
)
