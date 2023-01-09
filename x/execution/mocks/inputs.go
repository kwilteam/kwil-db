package mocks

import (
	execTypes "kwil/x/types/execution"
)

var (
	// it is important this corresponds to the queries in database.go

	// Insert1 and Insert2 (have 1 non-static)
	Insert1Inputs = []*execTypes.UserInput{&Param2Input}
	Insert2Inputs = []*execTypes.UserInput{&Param3Input}

	// Update1 (has 1 non-static)
	Update1Inputs = []*execTypes.UserInput{&Param2Input}

	// Update2 (has 2 non-static)
	Update2Inputs = []*execTypes.UserInput{&Param3Input, &Where1Input}

	// Delete1 (has 0 non-static)
	Delete1Inputs = []*execTypes.UserInput{}

	// Delete2 (has 1 non-static)
	Delete2Inputs = []*execTypes.UserInput{&Where1Input}

	Param2Input = execTypes.UserInput{
		Name:  Parameter2.Name,
		Value: "421",
	}

	// Insert2 (has 1 non-static)
	Param3Input = execTypes.UserInput{
		Name:  Parameter3.Name,
		Value: "true",
	}

	Where1Input = execTypes.UserInput{
		Name:  WhereClause1.Name,
		Value: "true",
	}
)
