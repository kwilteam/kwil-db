package pdb

type nameValidationContext struct {
	RelationNames       map[RelationIdentifier][]FieldID
	IndexNames          map[IndexName]struct{}
	UniqueNames         map[IndexName]struct{}
	PrimaryKeyNames     map[ModelID]string
	ConstraintNamespace ConstraintNamespace
}

func newNameValidationContext(db *Db) *nameValidationContext {
	relationNames := map[RelationIdentifier][]FieldID{}
	indexNames := map[IndexName]struct{}{}
	uniqueNames := map[IndexName]struct{}{}
	primaryKeyNames := map[ModelID]string{}

	for _, model := range db.WalkModels() {
		modelID := model.ID()
		for _, field := range model.RelationFields() {
			ident := RelationIdentifier{
				ModelA: field.Model().ID(),
				ModelB: field.RelatedModel().ID(),
				Name:   field.RelationName(),
			}
			relationNames[ident] = append(relationNames[ident], field.ID())
		}

		for _, index := range model.Indexes() {
			if name := index.Name(); name != "" {
				idxName := IndexName{Model: modelID, Name: name}
				if index.IsUnique() {
					uniqueNames[idxName] = struct{}{}
				} else {
					indexNames[idxName] = struct{}{}
				}
			}
		}

		if pk, ok := model.PrimaryKey(); ok && pk.Name() != "" {
			primaryKeyNames[modelID] = pk.Name()
		}
	}

	return &nameValidationContext{
		RelationNames:       relationNames,
		IndexNames:          indexNames,
		UniqueNames:         uniqueNames,
		PrimaryKeyNames:     primaryKeyNames,
		ConstraintNamespace: inferNamespaces(db),
	}
}

func inferNamespaces(db *Db) ConstraintNamespace {
	ns := ConstraintNamespace{
		Global:      map[GlobalConstraint]uint{},
		Local:       map[LocalConstraint]uint{},
		LocalCustom: map[LocalCustomConstraint]uint{},
	}

	for _, model := range db.WalkModels() {
		if pk, ok := model.PrimaryKey(); ok {
			ns.Local[LocalConstraint{Model: model.ID(), Name: pk.ConstraintName(), Scope: ModelPrimaryKeyKeyIndex}]++
			ns.LocalCustom[LocalCustomConstraint{Model: model.ID(), Name: pk.Name()}]++
		}
		for _, index := range model.Indexes() {
			ns.Local[LocalConstraint{Model: model.ID(), Name: index.ConstraintName(), Scope: ModelPrimaryKeyKeyIndex}]++
			ns.LocalCustom[LocalCustomConstraint{Model: model.ID(), Name: index.Name()}]++
		}
	}

	return ns
}
