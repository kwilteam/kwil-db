package serialize

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

func (s *Serializable) Action() (*dto.Action, error) {
	if s.Type != IdentifierAction {
		return nil, fmt.Errorf("cannot unserialize type '%s' as action", s.Type)
	}

	switch s.Version {
	case actionVersion:
		return actionVersion1ToAction(s.Data)
	default:
		return nil, fmt.Errorf("unsupported action version '%d'", s.Version)
	}
}

func SerializeAction(action *dto.Action) (*Serializable, error) {
	ser, err := convertActionToCurrentVersion(action)
	if err != nil {
		return nil, err
	}

	data, err := ser.Serialize()
	if err != nil {
		return nil, err
	}

	return &Serializable{
		Name:    action.Name,
		Type:    IdentifierAction,
		Version: actionVersion,
		Data:    data,
	}, nil
}

func convertActionToCurrentVersion(action *dto.Action) (serializer, error) {
	return &actionVersion1{
		Name:       action.Name,
		Inputs:     action.Inputs,
		Public:     action.Public,
		Statements: action.Statements,
	}, nil
}

func actionVersion1ToAction(data []byte) (*dto.Action, error) {
	var action actionVersion1
	err := json.Unmarshal(data, &action)
	if err != nil {
		return nil, err
	}

	return &dto.Action{
		Name:       action.Name,
		Inputs:     action.Inputs,
		Public:     action.Public,
		Statements: action.Statements,
	}, nil
}

type actionVersion1 struct {
	Name       string   `json:"name"`
	Inputs     []string `json:"inputs"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

func (a *actionVersion1) Serialize() ([]byte, error) {
	return json.Marshal(a)
}
