package builder

import "github.com/kwilteam/kwil-db/pkg/engine/tree"

func Builder() BuildSelector {
	return &builder{}
}

type builder struct{}

func (b *builder) Insert() TableSelector[InsertBuilderWithAlias] {
	return newInsert(tree.InsertTypeInsert)
}

func (b *builder) Replace() TableSelector[InsertBuilderWithAlias] {
	return newInsert(tree.InsertTypeReplace)
}

func (b *builder) InsertOrReplace() TableSelector[InsertBuilderWithAlias] {
	return newInsert(tree.InsertTypeInsertOrReplace)
}

type BuildSelector interface {
	Insert() TableSelector[InsertBuilderWithAlias]
	Replace() TableSelector[InsertBuilderWithAlias]
	InsertOrReplace() TableSelector[InsertBuilderWithAlias]
}
