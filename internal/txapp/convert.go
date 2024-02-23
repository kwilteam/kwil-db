package txapp

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func convertSchemaToEngine(old *transactions.Schema) (*common.Schema, error) {
	procedures, err := convertActionsToEngine(old.Actions)
	if err != nil {
		return nil, err
	}

	tables, err := convertTablesToEngine(old.Tables)
	if err != nil {
		return nil, err
	}

	extensions, err := convertExtensionsToEngine(old.Extensions)
	if err != nil {
		return nil, err
	}

	return &common.Schema{
		Name:       old.Name,
		Tables:     tables,
		Procedures: procedures,
		Extensions: extensions,
	}, nil
}

func convertTablesToEngine(tables []*transactions.Table) ([]*common.Table, error) {
	convTables := make([]*common.Table, len(tables))
	for i, table := range tables {
		columns, err := convertColumnsToEngine(table.Columns)
		if err != nil {
			return nil, err
		}

		indexes, err := convertIndexesToEngine(table.Indexes)
		if err != nil {
			return nil, err
		}

		foreignKeys, err := convertForeignKeysToEngine(table.ForeignKeys)
		if err != nil {
			return nil, err
		}

		convTables[i] = &common.Table{
			Name:        table.Name,
			Columns:     columns,
			Indexes:     indexes,
			ForeignKeys: foreignKeys,
		}
	}

	return convTables, nil
}

func convertColumnsToEngine(columns []*transactions.Column) ([]*common.Column, error) {
	convColumns := make([]*common.Column, len(columns))
	for i, column := range columns {
		colType := common.DataType(column.Type)
		if err := colType.Clean(); err != nil {
			return nil, err
		}

		attributes, err := convertAttributesToEngine(column.Attributes)
		if err != nil {
			return nil, err
		}

		convColumns[i] = &common.Column{
			Name:       column.Name,
			Type:       colType,
			Attributes: attributes,
		}
	}

	return convColumns, nil
}

func convertAttributesToEngine(attributes []*transactions.Attribute) ([]*common.Attribute, error) {
	convAttributes := make([]*common.Attribute, len(attributes))
	for i, attribute := range attributes {
		attrType := common.AttributeType(attribute.Type)
		if err := attrType.Clean(); err != nil {
			return nil, err
		}

		convAttributes[i] = &common.Attribute{
			Type:  attrType,
			Value: attribute.Value, // Assuming you have a function to parse the value based on its type
		}
	}

	return convAttributes, nil
}

func convertIndexesToEngine(indexes []*transactions.Index) ([]*common.Index, error) {
	convIndexes := make([]*common.Index, len(indexes))
	for i, index := range indexes {
		indexType := common.IndexType(index.Type)
		if err := indexType.Clean(); err != nil {
			return nil, err
		}

		convIndexes[i] = &common.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    indexType,
		}
	}

	return convIndexes, nil
}

func convertActionsToEngine(actions []*transactions.Action) ([]*common.Procedure, error) {
	convActions := make([]*common.Procedure, len(actions))
	for i, action := range actions {
		mods, err := convertModifiersToEngine(action.Mutability, action.Auxiliaries)
		if err != nil {
			return nil, err
		}

		convActions[i] = &common.Procedure{
			Name:        action.Name,
			Annotations: action.Annotations,
			Public:      action.Public,
			Modifiers:   mods,
			Args:        action.Inputs,
			Statements:  action.Statements,
		}
	}

	return convActions, nil
}

func convertModifiersToEngine(mutability string, auxiliaries []string) ([]common.Modifier, error) {
	mods := make([]common.Modifier, 0)
	switch strings.ToLower(mutability) {
	case transactions.MutabilityUpdate.String():
		break
	case transactions.MutabilityView.String():
		mods = append(mods, common.ModifierView)
	default:
		return nil, fmt.Errorf("unknown mutability type: %v", mutability)
	}

	for _, aux := range auxiliaries {
		switch strings.ToLower(aux) {
		case transactions.AuxiliaryTypeMustSign.String():
			mods = append(mods, common.ModifierAuthenticated)
		case transactions.AuxiliaryTypeOwner.String():
			mods = append(mods, common.ModifierOwner)
		default:
			return nil, fmt.Errorf("unknown auxiliary type: %v", aux)
		}
	}

	return mods, nil
}

func convertExtensionsToEngine(extensions []*transactions.Extension) ([]*common.Extension, error) {
	convExtensions := make([]*common.Extension, len(extensions))
	for i, extension := range extensions {
		convExtensions[i] = &common.Extension{
			Name:           extension.Name,
			Initialization: convertExtensionConfigToEngine(extension.Config),
			Alias:          extension.Alias,
		}
	}

	return convExtensions, nil
}

func convertExtensionConfigToEngine(configs []*transactions.ExtensionConfig) []*common.ExtensionConfig {
	convConfigs := make([]*common.ExtensionConfig, len(configs))
	for i, config := range configs {
		convConfigs[i] = &common.ExtensionConfig{
			Key:   config.Argument,
			Value: config.Value,
		}
	}

	return convConfigs
}

func convertForeignKeysToEngine(fks []*transactions.ForeignKey) ([]*common.ForeignKey, error) {
	results := make([]*common.ForeignKey, len(fks))
	for i, fk := range fks {
		actions := make([]*common.ForeignKeyAction, len(fk.Actions))
		for j, action := range fk.Actions {
			newAction := &common.ForeignKeyAction{
				On: common.ForeignKeyActionOn(action.On),
				Do: common.ForeignKeyActionDo(action.Do),
			}
			err := newAction.Clean()
			if err != nil {
				return nil, err
			}

			actions[j] = newAction
		}

		results[i] = &common.ForeignKey{
			ChildKeys:   fk.ChildKeys,
			ParentKeys:  fk.ParentKeys,
			ParentTable: fk.ParentTable,
			Actions:     actions,
		}
	}

	return results, nil
}
