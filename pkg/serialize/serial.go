/*
	The serialize package makes the old way of serializing / deserializing transaction payloads compatible
	with the refactored codebase.  This will likely be deleted, but it is going here now to isolate it.
*/

package serialize

import (
	"encoding/json"
	"fmt"

	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func DeserializeSchema(bts []byte) (*engineTypes.Schema, error) {
	schma := &Schema{}
	err := json.Unmarshal(bts, schma)
	if err != nil {
		return nil, err
	}

	return convertSchema(schma)
}

func DeserializeDBID(bts []byte) (string, error) {
	di := &DatasetIdentifier{}
	err := json.Unmarshal(bts, di)
	if err != nil {
		return "", err
	}

	return utils.GenerateDBID(di.Name, di.Owner), nil
}

func DeserializeActionPaload(payload []byte) (*tx.ExecuteActionPayload, error) {
	exec := tx.ExecuteActionPayload{}

	err := json.Unmarshal(payload, &exec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal action execution: %w", err)
	}

	return &exec, nil
}
