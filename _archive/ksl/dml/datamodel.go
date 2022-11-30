package dml

import "github.com/samber/mo"

type Datamodel struct {
	Enums  []*Enum
	Models []*Model
}

func NewDatamodel() *Datamodel {
	return &Datamodel{}
}

func (d *Datamodel) FindModel(name string) mo.Option[*Model] {
	for _, model := range d.Models {
		if model.Name == name {
			return mo.Some(model)
		}
	}
	return mo.None[*Model]()
}

func (d *Datamodel) FindEnum(name string) mo.Option[*Enum] {
	for _, enum := range d.Enums {
		if enum.Name == name {
			return mo.Some(enum)
		}
	}
	return mo.None[*Enum]()
}
