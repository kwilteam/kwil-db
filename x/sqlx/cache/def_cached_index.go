package cache

import (
	"fmt"
	spec "kwil/x/sqlx"
	"kwil/x/sqlx/models"
)

type Index struct {
	Name    string
	Table   string
	Columns []string
	Using   spec.IndexType
}

func (c *Index) From(m *models.Index) error {
	c.Name = m.Name
	c.Table = m.Table
	c.Columns = m.Columns
	using, err := spec.Conversion.ConvertIndex(m.Using)
	if err != nil {
		return fmt.Errorf("failed to convert index type: %s", err.Error())
	}
	c.Using = using
	return nil
}
