package engine

type Execution struct {
	TxID      string
	DatasetID string
	Procedure string
	Args      []map[string]any
	Caller    string
}

/*
func (e *Engine) BatchExecute(ctx context.Context, executions []*Execution) ([]any, error) {
	// TODO: 2pc
	for _, execution := range executions {
		dataset, ok := e.datasets[execution.DatasetID]
		if !ok {
			return nil, fmt.Errorf("dataset not found: %s", execution.DatasetID)
		}

		extension, ok := e.extensions[dataset.ExtensionName]
		if !ok {
			return nil, errors.New("extension not found")
		}

		extensionClient, err := extension.Initialize(ctx, dataset.Metadata)
		if err != nil {
			return nil, err
		}

		_, err = extensionClient.CallMethod(ctx, execution.Procedure, execution.Args...)
		if err != nil {
			return nil, err
		}
	}
}

type changeset struct {
	TxID      string
	DatasetID string
	Changes   []byte
}

func generateChangeset(dataset Dataset) {}
*/
