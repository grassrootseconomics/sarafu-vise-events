package event

import (
	"context"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"

	"git.defalsify.org/vise.git/persist"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
)

const (
	evReg = "CUSTODIAL_REGISTRATION"
	accountCreatedFlag = 9
)

// fields used for handling custodial registration event.
type eventCustodialRegistration struct {
	Account string
}

// attempt to coerce event as custodial registration.
func asCustodialRegistrationEvent(gev *geEvent.Event) (*eventCustodialRegistration, bool) {
	var ok bool
	var ev eventCustodialRegistration
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

// handle custodial registration.
//
// TODO: implement account created in userstore instead, so that
// the need for persister and state use here is eliminated (it
// introduces concurrency risks)
func handleCustodialRegistration(ctx context.Context, userStore *store.UserDataStore, pr *persist.Persister, ev *eventCustodialRegistration) error {
	identity, err := lookup.IdentityFromAddress(ctx, userStore, ev.Account)
	if err != nil {
		return err
	}
	err = pr.Load(identity.SessionId)
	if err != nil {
		return err
	}
	st := pr.GetState()
	st.SetFlag(accountCreatedFlag)
	return pr.Save(identity.SessionId)
}
