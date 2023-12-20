package types

import (
	"fmt"
	"strings"
)

type ForeignKey struct {
	// ChildKeys are the columns that are referencing another.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "a" is the child key
	ChildKeys []string `json:"child_keys"`

	// ParentKeys are the columns that are being referred to.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2.b" is the parent key
	ParentKeys []string `json:"parent_keys"`

	// ParentTable is the table that holds the parent columns.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2.b" is the parent table
	ParentTable string `json:"parent_table"`

	// Action refers to what the foreign key should do when the parent is altered.
	// This is NOT the same as a database action;
	// however sqlite's docs refer to these as actions,
	// so we should be consistent with that.
	// For example, ON DELETE CASCADE is a foreign key action
	Actions []*ForeignKeyAction `json:"actions"`
}

// Clean runs a set of validations and cleans the foreign key
func (f *ForeignKey) Clean() error {
	if len(f.ChildKeys) != len(f.ParentKeys) {
		return fmt.Errorf("foreign key must have same number of child and parent keys")
	}

	for _, action := range f.Actions {
		err := action.Clean()
		if err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdents(&f.ChildKeys),
		cleanIdents(&f.ParentKeys),
		cleanIdent(&f.ParentTable),
	)
}

// Copy returns a copy of the foreign key
func (f *ForeignKey) Copy() *ForeignKey {
	actions := make([]*ForeignKeyAction, len(f.Actions))
	for i, action := range f.Actions {
		actions[i] = action.Copy()
	}

	return &ForeignKey{
		ChildKeys:   f.ChildKeys,
		ParentKeys:  f.ParentKeys,
		ParentTable: f.ParentTable,
		Actions:     actions,
	}
}

// ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On ForeignKeyActionOn `json:"on"`

	// Do specifies what a foreign key action should do
	Do ForeignKeyActionDo `json:"do"`
}

// Clean runs a set of validations and cleans the attributes in ForeignKeyAction
func (f *ForeignKeyAction) Clean() error {
	return runCleans(
		f.On.Clean(),
		f.Do.Clean(),
	)
}

// Copy returns a copy of the foreign key action
func (f *ForeignKeyAction) Copy() *ForeignKeyAction {
	return &ForeignKeyAction{
		On: f.On,
		Do: f.Do,
	}
}

// ForeignKeyActionOn specifies when a foreign key action should occur.
// It can be either "UPDATE" or "DELETE".
type ForeignKeyActionOn string

const (
	// ON_UPDATE is used to specify an action should occur when a parent key is updated
	ON_UPDATE ForeignKeyActionOn = "UPDATE"
	// ON_DELETE is used to specify an action should occur when a parent key is deleted
	ON_DELETE ForeignKeyActionOn = "DELETE"
)

// IsValid checks whether or not the string is a valid ForeignKeyActionOn
func (f *ForeignKeyActionOn) IsValid() bool {
	upper := strings.ToUpper(f.String())
	return upper == ON_UPDATE.String() ||
		upper == ON_DELETE.String()
}

// Clean checks whether the string is valid, and will convert it to the correct case.
func (f *ForeignKeyActionOn) Clean() error {
	upper := strings.ToUpper(f.String())

	if !f.IsValid() {
		return fmt.Errorf("invalid ForeginKeyActionOn. received: %s", f.String())
	}

	*f = ForeignKeyActionOn(upper)

	return nil
}

// String returns the ForeignKeyActionOn as a string
func (f ForeignKeyActionOn) String() string {
	return string(f)
}

// ForeignKeyActionDo specifies what should be done when a foreign key action is triggered.
type ForeignKeyActionDo string

const (
	// DO_NO_ACTION does nothing when a parent key is altered
	DO_NO_ACTION ForeignKeyActionDo = "NO ACTION"

	// DO_RESTRICT prevents the parent key from being altered
	DO_RESTRICT ForeignKeyActionDo = "RESTRICT"

	// DO_SET_NULL sets the child key(s) to NULL
	DO_SET_NULL ForeignKeyActionDo = "SET NULL"

	// DO_SET_DEFAULT sets the child key(s) to their default values
	DO_SET_DEFAULT ForeignKeyActionDo = "SET DEFAULT"

	// DO_CASCADE updates the child key(s) or deletes the records (depending on the action type)
	DO_CASCADE ForeignKeyActionDo = "CASCADE"
)

// String returns the ForeignKeyActionDo as a string
func (f ForeignKeyActionDo) String() string {
	return string(f)
}

// IsValid checks if the string is a valid ForeignKeyActionDo
func (f *ForeignKeyActionDo) IsValid() bool {
	upper := strings.ToUpper(f.String())

	return upper == DO_NO_ACTION.String() ||
		upper == DO_RESTRICT.String() ||
		upper == DO_SET_NULL.String() ||
		upper == DO_SET_DEFAULT.String() ||
		upper == DO_CASCADE.String()
}

// Clean checks the validity or the string, and converts it to the correct case
func (f *ForeignKeyActionDo) Clean() error {
	upper := strings.ToUpper(f.String())

	if !f.IsValid() {
		return fmt.Errorf("invalid ForeignKeyActionDo. received: %s", upper)
	}

	*f = ForeignKeyActionDo(upper)

	return nil
}
