package pdb

import "ksl/syntax/ast"

func (db *Db) WalkModels() []ModelWalker {
	var models []ModelWalker
	for eid, entry := range db.Ast.Entries() {
		if _, ok := entry.(*ast.Model); ok {
			models = append(models, db.WalkModel(ModelID(eid)))
		}
	}
	return models
}

func (pdb *Db) WalkModel(id ModelID) ModelWalker {
	return ModelWalker{db: pdb, id: id}
}

func (db *Db) WalkEnums() []EnumWalker {
	var enums []EnumWalker
	for eid, entry := range db.Ast.Entries() {
		if _, ok := entry.(*ast.Enum); ok {
			enums = append(enums, db.WalkEnum(EnumID(eid)))
		}
	}
	return enums
}

func (db *Db) WalkEnum(id EnumID) EnumWalker {
	return EnumWalker{db: db, id: id}
}

func (db *Db) WalkRelations() []RelationWalker {
	var relations []RelationWalker
	for rid := range db.Relations.Storage {
		relations = append(relations, db.WalkRelation(RelationID(rid)))
	}
	return relations
}

func (db *Db) WalkRelation(id RelationID) RelationWalker {
	return RelationWalker{db: db, id: id}
}

func (db *Db) WalkCompleteInlineRelations() []CompleteInlineRelationWalker {
	var relations []CompleteInlineRelationWalker
	for _, item := range db.Relations.Relations() {
		if item.Relation.IsImplicitManyToMany() {
			continue
		}

		var fieldA, fieldB FieldID
		switch typ := item.Relation.Type.(type) {
		case OneToOneBoth:
			fieldA, fieldB = typ.FieldA, typ.FieldB
		case OneToManyBoth:
			fieldA, fieldB = typ.FieldA, typ.FieldB
		default:
			continue
		}
		relations = append(relations, db.WalkCompleteInlineRelation(item.ModelA, item.ModelB, fieldA, fieldB))
	}
	return relations
}

func (db *Db) WalkCompleteInlineRelation(modelA, modelB ModelID, fieldA, fieldB FieldID) CompleteInlineRelationWalker {
	return CompleteInlineRelationWalker{db: db, modelA: modelA, fieldA: fieldA, modelB: modelB, fieldB: fieldB}
}
