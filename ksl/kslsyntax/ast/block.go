package ast

import "ksl"

var _ Node = (*Block)(nil)

type Block struct {
	Type     *Str
	Name     *Str
	Modifier *Str
	Target   *Str
	Labels   *BlockLabels
	Body     *Body

	SrcRange ksl.Range
}

func (b *Block) GetLabel(key string) (*Attribute, bool) {
	if b == nil {
		return nil, false
	}
	return b.Labels.Label(key)
}

func (b *Block) MustLabel(key string) *Attribute {
	l, _ := b.GetLabel(key)
	return l
}

func (b *Block) GetType() string {
	if b == nil {
		return ""
	}
	return b.Type.GetString()
}

func (b *Block) GetName() string {
	if b == nil {
		return ""
	}
	return b.Name.GetString()
}

func (b *Block) HasModifier() bool {
	if b == nil {
		return false
	}
	return b.Modifier.GetString() != ""
}

func (b *Block) GetModifier() string {
	if b == nil {
		return ""
	}
	return b.Modifier.GetString()
}

func (b *Block) GetTarget() string {
	if b == nil {
		return ""
	}
	return b.Target.GetString()
}

func (b *Block) GetAllLabels() Attributes {
	if b == nil {
		return nil
	}
	return b.Labels.GetValues()
}

func (b *Block) GetAttributes() Attributes {
	if b == nil {
		return nil
	}
	return b.Body.GetAttributes()
}

func (b *Block) GetBlocks() Blocks {
	if b == nil {
		return nil
	}
	return b.Body.GetBlocks()
}

func (b *Block) GetAnnotations() Annotations {
	if b == nil {
		return nil
	}
	return b.Body.GetAnnotations()
}

func (b *Block) GetEnumValues() []string {
	if b == nil {
		return nil
	}
	return b.Body.GetEnumValues()
}

func (b *Block) GetDefinitions() Definitions {
	if b == nil {
		return nil
	}
	return b.Body.GetDefinitions()
}

type Blocks []*Block

func (els Blocks) ByType() map[string]Blocks {
	ret := make(map[string]Blocks)
	for _, el := range els {
		ty := el.Type.Value
		if ret[ty] == nil {
			ret[ty] = make(Blocks, 0, 1)
		}
		ret[ty] = append(ret[ty], el)
	}
	return ret
}
