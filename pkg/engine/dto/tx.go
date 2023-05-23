package dto

type TxContext struct {
	Caller  string
	Action  string
	Dataset string
}

// fillDefaults fills in default values for the options.
func (t *TxContext) fillDefaults() *TxContext {
	if t == nil {
		t = &TxContext{}
	}

	if t.Caller == "" {
		t.Caller = defaultCallerAddress
	}

	return t
}

// FillInputs adds the ExecOpts values to the inputs map.
func (t *TxContext) FillInputs(inputs map[string]any) map[string]any {
	t = t.fillDefaults()
	if inputs == nil {
		inputs = make(map[string]any)
	}

	inputs[callerVarName] = t.Caller
	inputs[actionVarName] = t.Action
	inputs[datasetVarName] = t.Dataset

	return inputs
}
