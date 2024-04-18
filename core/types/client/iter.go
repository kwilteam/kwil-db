package client

import "fmt"

// Records providers an iterator over a set of records.
type Records struct {
	// index tracks the current row index for the iterator.
	index int

	// rows is the underlying sql.Rows object.
	records []*Record
}

// Record represents a single row in a set of records.
type Record map[string]any

func newRecordFromMap(rec map[string]any) *Record {
	record := Record(rec)
	return &record
}

// NewRecords constructs a Records instance for iterating over a Record slice.
// DEPRECATED: This is intended for internal use. If you have all the records in
// a slice, you don't need to construct a Records iterator.
func NewRecords(records []*Record) *Records {
	return &Records{
		index:   -1,
		records: records,
	}
}

// NewRecordsFromMaps creates a Records from a slice of the maps of the same
// shape as an individual Record.
func NewRecordsFromMaps(recs []map[string]any) *Records {
	records := make([]*Record, len(recs))
	for i, rec := range recs {
		records[i] = newRecordFromMap(rec)
	}

	return NewRecords(records)
}

// Next steps to the next Record, returning false if there are no more records.
// Next must be used prior to accessing the first record with the Record method.
func (r *Records) Next() bool {
	r.index++

	if r.records == nil {
		return false
	}

	return r.index < len(r.records)
}

// Reset resets the iterator to the initial state. Use Next to get the first
// record.
func (r *Records) Reset() {
	r.index = -1
}

// Record returns the current Record. Use Next to iterate through the records.
func (r *Records) Record() *Record {
	if r.records == nil {
		return &Record{}
	}

	return r.records[r.index]
}

// Export returns all of the records in a slice. The map in each slice is
// equivalent to a Record, which is keyed by the column name.
func (r *Records) Export() []map[string]any {
	if r.records == nil {
		return make([]map[string]any, 0)
	}

	records := make([]map[string]any, len(r.records))

	for i, record := range r.records {
		records[i] = *record
	}

	return records
}

// ExportString is like Export, but the values in each map are converted to
// strings.
func (r *Records) ExportString() []map[string]string {
	if r.records == nil {
		return make([]map[string]string, 0)
	}

	records := make([]map[string]string, len(r.records))

	for i, record := range r.records {
		records[i] = record.String()
	}

	return records
}

// Map returns the record as a map. This is equivalent to map[string]any(r).
// This returns a reference to the underlying map represented by the Record.
func (r Record) Map() map[string]any {
	return r
}

// String converts the Record into a map with the values converted to strings.
func (r Record) String() map[string]string {
	rec := make(map[string]string)
	for k, v := range r {
		rec[k] = fmt.Sprintf("%v", v)
	}

	return rec
}
