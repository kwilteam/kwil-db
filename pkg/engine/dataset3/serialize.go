package dataset2

/*
type serliazableMetadata struct {
	Type serializeableType
	Data []byte
}

type serializeableType byte

const (
	serializeableTypeProcedure serializeableType = iota
	serializeableTypeExtension
)

func serializeOperation(op Operation) ([]byte, error) {
	var bts []byte
	var err error
	switch oper := op.(type) {
	case *DMLOperation:
		bts, err = json.Marshal(oper)
	case *ExtensionMethodOperation:
		bts, err = json.Marshal(oper)
	case *ProcedureCallOperation:
		bts, err = json.Marshal(oper)
	default:
		return nil, fmt.Errorf("unknown operation type %T", op)
	}
	if err != nil {
		return nil, err
	}

	return append([]byte{byte(op.Type())}, bts...), nil
}

func deserializeOperation(bts []byte) (Operation, error) {
	if len(bts) == 0 {
		return nil, fmt.Errorf("cannot deserialize empty byte array")
	}

	var op Operation
	var err error
	switch OperationType(bts[0]) {
	case OperationTypeDML:
		var dml DMLOperation
		err = json.Unmarshal(bts[1:], &op)
		op = &dml
	case OperationTypeExtensionMethod:
		var ext ExtensionMethodOperation
		err = json.Unmarshal(bts[1:], &op)
		op = &ext
	case OperationTypeProcedureCall:
		var proc ProcedureCallOperation
		err = json.Unmarshal(bts[1:], &op)
		op = &proc
	default:
		return nil, fmt.Errorf("unknown operation type %d", bts[0])
	}

	return op, err
}

// serialized
type serializedProcedure struct {
	Procedure []byte   `json:"procedure"`
	Body      [][]byte `json:"data"`
}

func serializeProcedure(proc *Procedure) ([]byte, error) {
	bts, err := json.Marshal(proc)
	if err != nil {
		return nil, err
	}

	var serializedOperations serializeableDualByteArray
	for _, op := range proc.Body {
		opBytes, err := serializeOperation(op)
		if err != nil {
			return nil, err
		}

		serializedOperations.Data = append(serializedOperations.Data, opBytes)
	}

	operationBytes, err := json.Marshal(serializedOperations)
	if err != nil {
		return nil, err
	}

	return append([]byte{byte(serializeableTypeProcedure)}, append(bts, operationBytes...)...), nil
}

func deserializeProcedure(bts []byte) (*StoredProcedure, error) {
	if len(bts) == 0 {
		return nil, fmt.Errorf("cannot deserialize empty byte array")
	}

	var proc StoredProcedure
	err := json.Unmarshal(bts[1:], &proc)
	if err != nil {
		return nil, err
	}

	return &proc, nil
}
*/
