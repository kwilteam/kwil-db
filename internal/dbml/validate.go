package dbml

type Validator interface {
	Validate(*DBML) error
}

type ValidatorFunc func(*DBML) error

func (f ValidatorFunc) Validate(dbml *DBML) error { return f(dbml) }

type ValidatorGroup struct {
	validators []Validator
}

func NewValidatorGroup(validators ...Validator) *ValidatorGroup {
	return &ValidatorGroup{validators: validators}
}

func (g *ValidatorGroup) Validate(dbml *DBML) error {
	for _, v := range g.validators {
		if err := v.Validate(dbml); err != nil {
			return err
		}
	}
	return nil
}
