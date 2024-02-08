package types

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/utils"
)

type Schema struct {
	Name string `json:"name"`
	// Owner is the identifier (generally an address in bytes or public key) of the owner of the schema
	Owner      []byte       `json:"owner"`
	Extensions []*Extension `json:"extensions"`
	Tables     []*Table     `json:"tables"`
	Procedures []*Procedure `json:"procedures"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (s *Schema) Clean() error {
	err := cleanIdent(&s.Name)
	if err != nil {
		return err
	}

	tableSet := make(map[string]struct{})
	for _, table := range s.Tables {
		err := table.Clean()
		if err != nil {
			return err
		}

		_, ok := tableSet[table.Name]
		if ok {
			return fmt.Errorf(`duplicate table name: "%s"`, table.Name)
		}

		tableSet[table.Name] = struct{}{}
	}

	procedureSet := make(map[string]struct{})
	for _, action := range s.Procedures {
		err := action.Clean()
		if err != nil {
			return err
		}

		_, ok := procedureSet[action.Name]
		if ok {
			return fmt.Errorf(`duplicate procedure name: "%s"`, action.Name)
		}

		procedureSet[action.Name] = struct{}{}
	}

	extensionSet := make(map[string]struct{})
	for _, extension := range s.Extensions {
		err := extension.Clean()
		if err != nil {
			return err
		}

		_, ok := extensionSet[extension.Alias]
		if ok {
			return fmt.Errorf(`duplicate extension alias: "%s"`, extension.Alias)
		}

		extensionSet[extension.Alias] = struct{}{}
	}

	return nil
}

func (s *Schema) DBID() string {
	return utils.GenerateDBID(s.Name, s.Owner)
}
