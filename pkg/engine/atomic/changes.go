package atomic

// changeTracker is an interface for a type that tracks changes that were made to the database
// should this be a change applier?  essentially deciding if it should immediately commit or write to WAL
type changeTracker interface {
	// TrackChange tracks a change that was made to the database
	TrackChange(change *change)
}

// nilChangeTracker is a change tracker that does nothing
// it is used when changes are not being tracked
type nilChangeTracker struct{}

func (n *nilChangeTracker) TrackChange(_ *change) {}

var _ changeTracker = (*nilChangeTracker)(nil)

type change struct {
	// ID is a unique identifier for the change
	ID []byte

	// DBID is the id of the dataset that the change is targeting
	DBID string

	// Type is the type of change
	Type changeType

	// Data is the data for the change
	Data []byte
}

type changeType uint8

const (
	ctCreateDataset changeType = iota
	ctDeleteDataset
	ctCreateTable
	ctExecuteStatement
)
