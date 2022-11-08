package ksl

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type DiagnosticCode string
type DiagnosticSeverity int

const (
	// DiagInvalid is the invalid zero value of DiagnosticSeverity
	DiagInvalid DiagnosticSeverity = iota

	// DiagError indicates that the problem reported by a diagnostic prevents
	// further progress in parsing and/or evaluating the subject.
	DiagError

	// DiagWarning indicates that the problem reported by a diagnostic warrants
	// user attention but does not prevent further progress. It is most
	// commonly used for showing deprecation notices.
	DiagWarning
)

// Diagnostic represents information to be presented to a user about an
// error or anomaly in parsing or evaluating configuration.
type Diagnostic struct {
	Severity DiagnosticSeverity

	Summary string
	Detail  string

	// Subject and Context are both source ranges relating to the diagnostic.
	//
	// Subject is a tight range referring to exactly the construct that
	// is problematic, while Context is an optional broader range (which should
	// fully contain Subject) that ought to be shown around Subject when
	// generating isolated source-code snippets in diagnostic messages.
	// If Context is nil, the Subject is also the Context.
	Subject *Range
	Context *Range
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s; %s", d.Subject, d.Summary, d.Detail)
}

type Diagnostics []*Diagnostic

func (d Diagnostics) Error() string {
	count := len(d)
	switch {
	case count == 0:
		return "no diagnostics"
	case count == 1:
		return d[0].Error()
	default:
		return multierror.Append(nil, d.Errs()...).Error()
	}
}

func (d Diagnostics) HasErrors() bool {
	for _, diag := range d {
		if diag.Severity == DiagError {
			return true
		}
	}
	return false
}

func (d Diagnostics) Errs() []error {
	var errs []error
	for _, diag := range d {
		if diag.Severity == DiagError {
			errs = append(errs, diag)
		}
	}

	return errs
}

type DiagnosticWriter interface {
	WriteDiagnostic(*Diagnostic) error
	WriteDiagnostics(Diagnostics) error
}
