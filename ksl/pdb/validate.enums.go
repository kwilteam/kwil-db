package pdb

import (
	"fmt"
	"ksl"
)

func (v *ValidationContext) runEnumValidators(enum EnumWalker) {
	for _, validator := range v.enums {
		v.diag(validator(enum)...)
	}
}

func (v *ValidationContext) validateEnumDatabaseNameClashes(db *Db) {
	dbnames := map[string]EnumID{}

	for _, enum := range db.WalkEnums() {
		key := enum.DatabaseName()
		if eid, ok := dbnames[key]; ok {
			v.diag(&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Duplicate database name",
				Detail:   fmt.Sprintf("Enum %q has a database name clash with enum %q", enum.Name(), db.Ast.GetEnum(eid).GetName()),
				Subject:  enum.AstEnum().Range().Ptr(),
			})
		}
		dbnames[key] = enum.ID()
	}
}

func (v *ValidationContext) validateEnumHasValues(enum EnumWalker) {
	if len(enum.Values()) == 0 {
		v.diag(&ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  "Empty enum",
			Detail:   fmt.Sprintf("Enum %q has no values", enum.Name()),
			Subject:  enum.AstEnum().Range().Ptr(),
		})
	}
}
