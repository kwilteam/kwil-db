package serialize

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

func (s *Serializable) Extension() (*dto.ExtensionInitialization, error) {
	if s.Type != IdentifierExtension {
		return nil, fmt.Errorf("cannot unserialize type '%s' as extension", s.Type)
	}

	switch s.Version {
	case extensionVersion:
		return extensionVersion1ToExtension(s.Data)
	default:
		return nil, fmt.Errorf("unsupported extension version '%d'", s.Version)
	}
}

func SerializeExtension(ext *dto.ExtensionInitialization) (*Serializable, error) {
	ser, err := convertExtensionToCurrentVersion(ext)
	if err != nil {
		return nil, err
	}

	data, err := ser.Serialize()
	if err != nil {
		return nil, err
	}

	return &Serializable{
		Name:    ext.Name,
		Type:    IdentifierExtension,
		Version: extensionVersion,
		Data:    data,
	}, nil
}

func convertExtensionToCurrentVersion(ext *dto.ExtensionInitialization) (serializer, error) {
	return &extensionVersion1{
		Name:     ext.Name,
		Metadata: ext.Metadata,
	}, nil
}

func extensionVersion1ToExtension(data []byte) (*dto.ExtensionInitialization, error) {
	var ext extensionVersion1
	err := json.Unmarshal(data, &ext)
	if err != nil {
		return nil, err
	}

	return &dto.ExtensionInitialization{
		Name:     ext.Name,
		Metadata: ext.Metadata,
	}, nil
}

type extensionVersion1 struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
}

func (e *extensionVersion1) Serialize() ([]byte, error) {
	return json.Marshal(e)
}
