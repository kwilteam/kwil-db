package types

type Schema struct {
	Name string
	// Owner is the hex encoded public key of the owner of the dataset
	Owner      string
	Extensions []*Extension
	Tables     []*Table
	Procedures []*Procedure
}

func (s *Schema) Clean() error {
	for _, table := range s.Tables {
		err := table.Clean()
		if err != nil {
			return err
		}
	}

	for _, action := range s.Procedures {
		err := action.Clean()
		if err != nil {
			return err
		}
	}

	return nil
}
