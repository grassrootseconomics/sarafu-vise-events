package event

import (
	apievent "git.grassecon.net/grassrootseconomics/sarafu-api/event"
	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"
)

const (
	evReg = apievent.EventRegistrationTag
	//accountCreatedFlag = 9
)

// attempt to coerce event as custodial registration.
func asCustodialRegistrationEvent(gev *geEvent.Event) (*apievent.EventCustodialRegistration, bool) {
	var ok bool
	var ev apievent.EventCustodialRegistration
	if gev.TxType != evReg {
		return nil, false
	}
	pl := gev.Payload
	ev.Account, ok = pl["account"].(string)
	if !ok {
		return nil, false
	}
	logg.Debugf("parsed ev", "pl", pl, "ev", ev)
	return &ev, true
}
