package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/hcl"
	"github.com/kwilteam/kwil-db/internal/schema"
	"github.com/kwilteam/kwil-db/internal/sqlx"
)

func attr(typ *hcl.Type, key string) (*hcl.Attr, bool) {
	for _, a := range typ.Attrs {
		if a.K == key {
			return a, true
		}
	}
	return nil, false
}

func hasEnumType(c *schema.Column) (*schema.EnumType, bool) {
	switch t := c.Type.Type.(type) {
	case *schema.EnumType:
		return t, true
	case *ArrayType:
		if e, ok := t.Type.(*schema.EnumType); ok {
			return e, true
		}
	}
	return nil, false
}

func identity(attrs []schema.Attr) (*Identity, bool) {
	i := &Identity{}
	if !schema.Has(attrs, i) {
		return nil, false
	}
	if i.Generation == "" {
		i.Generation = defaultIdentityGen
	}
	if i.Sequence == nil {
		i.Sequence = &Sequence{Start: defaultSeqStart, Increment: defaultSeqIncrement}
		return i, true
	}
	if i.Sequence.Start == 0 {
		i.Sequence.Start = defaultSeqStart
	}
	if i.Sequence.Increment == 0 {
		i.Sequence.Increment = defaultSeqIncrement
	}
	return i, true
}

// formatPartition returns the string representation of the
// partition key according to the PostgreSQL format/grammar.
func formatPartition(p Partition) (string, error) {
	b := &schema.Builder{QuoteChar: '"'}
	b.P("PARTITION BY")
	switch t := strings.ToUpper(p.T); t {
	case PartitionTypeRange, PartitionTypeList, PartitionTypeHash:
		b.P(t)
	default:
		return "", fmt.Errorf("unknown partition type: %q", t)
	}
	if len(p.Parts) == 0 {
		return "", errors.New("missing parts for partition key")
	}
	b.Wrap(func(b *schema.Builder) {
		b.MapComma(p.Parts, func(i int, b *schema.Builder) {
			switch k := p.Parts[i]; {
			case k.C != nil:
				b.Ident(k.C.Name)
			case k.X != nil:
				b.P(sqlx.MayWrap(k.X.(*schema.RawExpr).X))
			}
		})
	})
	return b.String(), nil
}
