package dto

// this file contains global variables and their defaults, like @caller

const (
	defaultCallerAddress = "0x0000000000000000000000000000000000000000"
	callerVarName        = "@caller"

	actionVarName = "@action"
	defaultAction = "_no_action_"

	datasetVarName = "@dataset"
	datasetDefault = "x00000000000000000000000000000000000000000000000000000000"
)

var GlobalVars = map[string]interface{}{
	callerVarName:  defaultCallerAddress,
	actionVarName:  defaultAction,
	datasetVarName: datasetDefault,
}
