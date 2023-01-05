package mocks

import (
	"kwil/x/execution/dto"
)

var (
	// it is important this corresponds to the queries in database.go

	// Insert1 and Insert2 (have 1 non-static)
	Insert1Inputs = []*dto.UserInput{&Param2Input}
	Insert2Inputs = []*dto.UserInput{&Param3Input}

	// Update1 (has 1 non-static)
	Update1Inputs = []*dto.UserInput{&Param2Input}

	// Update2 (has 2 non-static)
	Update2Inputs = []*dto.UserInput{&Param3Input, &Where1Input}

	// Delete1 (has 0 non-static)
	Delete1Inputs = []*dto.UserInput{}

	// Delete2 (has 1 non-static)
	Delete2Inputs = []*dto.UserInput{&Where1Input}

	Param2Input = dto.UserInput{
		Name:  Parameter2.Name,
		Value: "421",
	}

	// Insert2 (has 1 non-static)
	Param3Input = dto.UserInput{
		Name:  Parameter3.Name,
		Value: "true",
	}

	Where1Input = dto.UserInput{
		Name:  WhereClause1.Name,
		Value: "true",
	}
)
