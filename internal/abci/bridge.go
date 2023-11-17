package abci

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/tokenbridge"
)

// bridgeEvents must implement BridgeEventsModule

var _ BridgeEventsModule = (*dummyBridgeEvents)(nil)

type testDep struct {
	eventID string
	amt     string
	acct    string
}

var testDepsToReport = []testDep{
	{
		eventID: "asdf",
		amt:     "12341234000000",
		acct:    "0xabababababab",
	},
}

type depositEvent struct {
	Account      string
	Amount       string
	Attestations []string // the attester identities, including self
}

var depositEvents = map[string]depositEvent{}

type dummyBridgeEvents struct {
	me string
}

func (be *dummyBridgeEvents) DepositsToReport(context.Context, string) (eventID, amt, acct []string, err error) {
	for i := range testDepsToReport {
		eventID = append(eventID, testDepsToReport[i].eventID)
		amt = append(amt, testDepsToReport[i].amt)
		acct = append(acct, testDepsToReport[i].acct)

		depositEvents[testDepsToReport[i].eventID] = depositEvent{
			Account:      testDepsToReport[i].acct,
			Amount:       testDepsToReport[i].amt,
			Attestations: []string{be.me},
		}
	}
	testDepsToReport = nil
	return
}

func (be *dummyBridgeEvents) AddDeposit(ctx context.Context, eventID, amt, acct, observer string) error {
	return nil
}

func (be *dummyBridgeEvents) DepositEvents(ctx context.Context, vals map[string]bool) (map[string]*tokenbridge.DepositEvent, error) {
	out := make(map[string]*tokenbridge.DepositEvent)
	for eid, de := range depositEvents {
		out[eid] = &tokenbridge.DepositEvent{
			EventID:      eid,
			Sender:       de.Account,
			Amount:       de.Amount,
			Attestations: de.Attestations,
		}
	}
	return out, nil
}

func (be *dummyBridgeEvents) MarkDepositActuated(ctx context.Context, eventID string) error {
	delete(depositEvents, eventID)
	return nil
}

// func (be *dummyBridgeEvents) HasThresholdAttestations(ctx context.Context, eventID string, validators map[string]bool) (bool, error) {
// 	return len(depositEvents) >= tokenbridge.Threshold(len(validators)), nil
// }
