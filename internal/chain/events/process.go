package events

import (
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

func (ef *EventFeed) ProcessLog(vLog ethTypes.Log) error {
	// Parse the event
	ev, err := ef.parseEvent(vLog)
	if err != nil {
		ef.log.Error().Err(err).Msg("error parsing event")
	}

	return ef.ProcessEvent(ev)
}

func (ef *EventFeed) ProcessEvent(ev Event) error {
	switch ev.GetName() {
	case "Deposit":
		return ef.ProcessDeposit(ev.(*DepositEvent))
	}

	return nil
}

func (ef *EventFeed) ProcessDeposit(ev *DepositEvent) error {
	err := ef.ds.Deposit(ev.Data.Amount, ev.Data.Caller.String(), ev.Tx, ev.Height)
	if err != nil {
		return err
	} else {
		ef.log.Info().Msgf("deposited to %s", ev.Data.Target)
	}
	return nil
}
