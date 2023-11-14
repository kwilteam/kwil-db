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
	Receiver []byte // this is the address of the receiver of the event
}
type DepositEvent struct {
	// Version   int
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Amount    string `json:"amount"`
	Height    int64  `json:"height"`
	TxHash    string `json:"txHash"`
	BlockHash string `json:"blockHash"`
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
