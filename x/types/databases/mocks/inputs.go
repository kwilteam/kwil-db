package mocks

import (
	execTypes "kwil/x/types/execution"
)

var (
	// it is important this corresponds to the queries in kwild.go

	// Insert1 and Insert2 (have 1 non-static)
	Insert1Inputs = []*execTypes.UserInput[[]byte]{&Param2Input}
	Insert2Inputs = []*execTypes.UserInput[[]byte]{&Param3Input}

	// Update1 (has 1 non-static)
	Update1Inputs = []*execTypes.UserInput[[]byte]{&Param2Input}

	// Update2 (has 2 non-static)
	Update2Inputs = []*execTypes.UserInput[[]byte]{&Param3Input, &Where1Input}

	// Delete1 (has 0 non-static)
	Delete1Inputs = []*execTypes.UserInput[[]byte]{}

	// Delete2 (has 1 non-static)
	Delete2Inputs = []*execTypes.UserInput[[]byte]{&Where1Input}

	Param2Input = execTypes.UserInput[[]byte]{
		Name:  Parameter2.Name,
		Value: []byte{3, 164, 1, 0, 0},
	}

	// Insert2 (has 1 non-static)
	Param3Input = execTypes.UserInput[[]byte]{
		Name:  Parameter3.Name,
		Value: []byte{5, 1},
	}

	Where1Input = execTypes.UserInput[[]byte]{
		Name:  WhereClause1.Name,
		Value: []byte{5, 1},
	}
)
