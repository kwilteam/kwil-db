package datasets

import (
	"github.com/kwilteam/kwil-db/internal/entity"
	engineDto "github.com/kwilteam/kwil-db/pkg/engine/dto"
)

func convertActions(actions []*engineDto.Action) []*entity.Action {
	entityActions := make([]*entity.Action, len(actions))
	for i, action := range actions {
		entityActions[i] = &entity.Action{
			Name:       action.Name,
			Inputs:     action.Inputs,
			Public:     action.Public,
			Statements: action.Statements,
		}
	}

	return entityActions
}

func convertTables(tables []*engineDto.Table) []*entity.Table {
	entityTables := make([]*entity.Table, len(tables))
	for i, table := range tables {
		entityTables[i] = &entity.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
	}

	return entityTables
}

func convertColumns(columns []*engineDto.Column) []*entity.Column {
	entityColumns := make([]*entity.Column, len(columns))
	for i, column := range columns {
		entityColumns[i] = &entity.Column{
			Name:       column.Name,
			Type:       column.Type.String(),
			Attributes: convertAttributes(column.Attributes),
		}
	}

	return entityColumns
}

func convertAttributes(attributes []*engineDto.Attribute) []*entity.Attribute {
	entityAttributes := make([]*entity.Attribute, len(attributes))
	for i, attribute := range attributes {
		entityAttributes[i] = &entity.Attribute{
			Type:  attribute.Type.String(),
			Value: attribute.Value,
		}
	}

	return entityAttributes
}

func convertIndexes(indexes []*engineDto.Index) []*entity.Index {
	entityIndexes := make([]*entity.Index, len(indexes))
	for i, index := range indexes {
		entityIndexes[i] = &entity.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type.String(),
		}
	}

	return entityIndexes
}

func convertActionsToDto(actions []*entity.Action) ([]*engineDto.Action, error) {
	entityActions := make([]*engineDto.Action, len(actions))
	for i, action := range actions {
		entityActions[i] = &engineDto.Action{
			Name:       action.Name,
			Inputs:     action.Inputs,
			Public:     action.Public,
			Statements: action.Statements,
		}
	}

	for i := range entityActions {
		err := entityActions[i].Clean()
		if err != nil {
			return nil, err
		}
	}

	return entityActions, nil
}

func convertTablesToDto(tables []*entity.Table) ([]*engineDto.Table, error) {
	entityTables := make([]*engineDto.Table, len(tables))
	for i, table := range tables {
		entityTables[i] = &engineDto.Table{
			Name:        table.Name,
			Columns:     convertColumnsToDto(table.Columns),
			Indexes:     convertIndexesToDto(table.Indexes),
			ForeignKeys: convertForeignKeysToDto(table.ForeignKeys),
		}
	}

	for i := range entityTables {
		err := entityTables[i].Clean()
		if err != nil {
			return nil, err
		}
	}

	return entityTables, nil
}

func convertForeignKeysToDto(foreignKeys []*entity.ForeignKey) []*engineDto.ForeignKey {
	entityForeignKeys := make([]*engineDto.ForeignKey, len(foreignKeys))
	for i, foreignKey := range foreignKeys {
		entityForeignKeys[i] = &engineDto.ForeignKey{
			ChildKeys:   foreignKey.ChildKeys,
			ParentKeys:  foreignKey.ParentKeys,
			ParentTable: foreignKey.ParentTable,
			Actions:     convertForeignKeyActionsToDto(foreignKey.Actions),
		}
	}

	return entityForeignKeys
}

func convertForeignKeyActionsToDto(foreignKeyActions []*entity.ForeignKeyAction) []*engineDto.ForeignKeyAction {
	entityForeignKeyActions := make([]*engineDto.ForeignKeyAction, len(foreignKeyActions))
	for i, foreignKeyAction := range foreignKeyActions {
		entityForeignKeyActions[i] = &engineDto.ForeignKeyAction{
			On: engineDto.ForeignKeyActionOn(foreignKeyAction.On),
			Do: engineDto.ForeignKeyActionDo(foreignKeyAction.Do),
		}
	}

	return entityForeignKeyActions
}

func convertColumnsToDto(columns []*entity.Column) []*engineDto.Column {
	entityColumns := make([]*engineDto.Column, len(columns))
	for i, column := range columns {
		entityColumns[i] = &engineDto.Column{
			Name:       column.Name,
			Type:       engineDto.DataType(column.Type),
			Attributes: convertAttributesToDto(column.Attributes),
		}
	}

	return entityColumns
}

func convertAttributesToDto(attributes []*entity.Attribute) []*engineDto.Attribute {
	entityAttributes := make([]*engineDto.Attribute, len(attributes))
	for i, attribute := range attributes {
		entityAttributes[i] = &engineDto.Attribute{
			Type:  engineDto.AttributeType(attribute.Type),
			Value: attribute.Value,
		}
	}

	return entityAttributes
}

func convertIndexesToDto(indexes []*entity.Index) []*engineDto.Index {
	entityIndexes := make([]*engineDto.Index, len(indexes))
	for i, index := range indexes {
		entityIndexes[i] = &engineDto.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    engineDto.IndexType(index.Type),
		}
	}

	return entityIndexes
}
