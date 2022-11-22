package schema

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func readYaml(bts []byte) (*Database[KwilType, KwilConstraint, KwilIndex], error) {
	var db Database[KwilType, KwilConstraint, KwilIndex]
	err := yaml.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func convertToPG(db *Database[KwilType, KwilConstraint, KwilIndex]) (*Database[PGType, PGConstraint, PGIndex], error) {
	var pgDB Database[PGType, PGConstraint, PGIndex]
	pgDB.Owner = db.Owner
	pgDB.Name = db.Name
	pgDB.DefaultRole = db.DefaultRole
	pgDB.Tables = make(map[string]Table[PGType, PGConstraint])
	pgDB.Roles = db.Roles
	pgDB.Queries = db.Queries
	for name, table := range db.Tables {
		var pgTable Table[PGType, PGConstraint]
		pgTable.Columns = make(map[string]Column[PGType, PGConstraint])
		for colName, col := range table.Columns {
			var pgCol Column[PGType, PGConstraint]
			pgt := col.Type.ToPG()
			if pgt == PGUnknownType {
				return nil, fmt.Errorf("unknown type: %s", col.Type)
			}
			pgCol.Type = pgt
			pgConsts := make([]PGConstraint, len(col.Constraints))
			for k, constraint := range col.Constraints {
				c := constraint.ToPG()
				if c == PGUnknownConstraint {
					return nil, fmt.Errorf("unknown constraint %s", constraint)
				}
				pgConsts[k] = c
			}
			pgCol.Constraints = pgConsts
			pgTable.Columns[colName] = pgCol
		}
		pgDB.Tables[name] = pgTable
	}
	pgDB.Indexes = make(map[string]Index[PGIndex])
	for name, index := range db.Indexes {
		var pgIndex Index[PGIndex]
		pgIndex.Table = index.Table
		pgIndex.Column = index.Column
		u := index.Using.ToPG()
		if u == PGUnknownIndex {
			return nil, fmt.Errorf("unknown index %s", index.Using)
		}
		pgIndex.Using = u
		pgDB.Indexes[name] = pgIndex
	}
	return &pgDB, nil
}

/*
func convertToKwil(db *Database[PGType, PGConstraint, PGIndex]) (*Database[KwilType, KwilConstraint, KwilIndex], error) {
	var kwilDB Database[KwilType, KwilConstraint, KwilIndex]
	kwilDB.Owner = db.Owner
	kwilDB.Name = db.Name
	kwilDB.DefaultRole = db.DefaultRole
	kwilDB.Tables = make([]Table[KwilType, KwilConstraint], len(db.Tables))
	kwilDB.Roles = db.Roles
	kwilDB.Queries = db.Queries
	for i, table := range db.Tables {
		kwilDB.Tables[i].Name = table.Name
		kwilDB.Tables[i].Columns = make([]Column[KwilType, KwilConstraint], len(table.Columns))
		for j, column := range table.Columns {
			kwilDB.Tables[i].Columns[j].Name = column.Name
			t := column.Type.ToKwil()
			if t == KwilUnknownType {
				return nil, fmt.Errorf("unknown type %s", column.Type)
			}
			kwilDB.Tables[i].Columns[j].Type = t
			kwilDB.Tables[i].Columns[j].Constraints = make([]KwilConstraint, len(column.Constraints))
			for k, constraint := range column.Constraints {
				c := constraint.ToKwil()
				if c == KwilUnknownConstraint {
					return nil, fmt.Errorf("unknown constraint %s", constraint)
				}
				kwilDB.Tables[i].Columns[j].Constraints[k] = c
			}
		}
	}
	kwilDB.Indexes = make([]Index[KwilIndex], len(db.Indexes))
	for i, index := range db.Indexes {
		kwilDB.Indexes[i].Name = index.Name
		kwilDB.Indexes[i].Table = index.Table
		kwilDB.Indexes[i].Column = index.Column
		u := index.Using.ToKwil()
		if u == KwilUnknownIndex {
			return nil, fmt.Errorf("unknown index %s", index.Using)
		}
		kwilDB.Indexes[i].Using = u
	}
	return &kwilDB, nil
}
*/
