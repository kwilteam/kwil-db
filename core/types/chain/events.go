package chain

type EventType int

const (
	Deposits EventType = iota
	Unknown
)

type Event struct {
	// Version  int
	ID       string
	Type     EventType
	Data     []byte
	Observer []byte // this is the address of the receiver of the event
}
type DepositEvent struct {
	// Version   int
	ID     string `json:"id"`
	Sender string `json:"sender"`
	Amount string `json:"amount"`
}

// eventID: Hash(DepositEvent)

func (ev EventType) String() string {
	switch ev {
	case Deposits:
		return "Deposit"
	default:
		return "Unknown"
	}
}
