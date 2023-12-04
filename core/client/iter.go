package client

import "fmt"

// Records providers an iterator over a set of records.
type Records struct {
	// index tracks the current row index for the iterator.
	index int

	// rows is the underlying sql.Rows object.
	records []*Record
}

type Record map[string]any

func NewRecordFromMap(rec map[string]any) *Record {
	record := Record(rec)
	return &record
}

func NewRecords(records []*Record) *Records {
	return &Records{
		index:   -1,
		records: records,
	}
}

func NewRecordsFromMaps(recs []map[string]any) *Records {
	records := make([]*Record, len(recs))
	for i, rec := range recs {
		records[i] = NewRecordFromMap(rec)
	}

	return NewRecords(records)
}

func (r *Records) Next() bool {
	r.index++

	if r.records == nil {
		return false
	}

	return r.index < len(r.records)
}

func (r *Records) Reset() {
	r.index = -1
}

func (r *Records) Record() *Record {
	if r.records == nil {
		return &Record{}
	}

	return r.records[r.index]
}

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

func (r Record) Map() map[string]any {
	return r
}

func (r Record) String() map[string]string {
	rec := make(map[string]string)
	for k, v := range r {
		rec[k] = fmt.Sprintf("%v", v)
	}

	return rec
}
