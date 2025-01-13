package event

import (
	"context"
	"fmt"
	"strconv"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"

	"git.grassecon.net/grassrootseconomics/common/hex"
	apievent "git.grassecon.net/grassrootseconomics/sarafu-api/event"
)

const (
	evTokenTransfer = apievent.EventTokenTransferTag
	// TODO: export from visedriver storage package
	//DATATYPE_USERSUB = 64
)

// formatter for transaction data
//
// TODO: current formatting is a placeholder.
func formatTransaction(tag string, idx int, item any) string {
	if tag == apievent.EventTokenTransferTag {
		tx, ok := item.(dataserviceapi.Last10TxResponse)
		if !ok {
			logg.Errorf("invalid formatting object", "tag", tag)
			return ""
		}
		return fmt.Sprintf("%d %s %s", idx, tx.DateBlock, tx.TxHash[:10])
	}
	logg.Warnf("unhandled formatting object", "tag", tag)
	return ""
}


// waiter to check whether object is available on dependency endpoints.
func updateWait(ctx context.Context) error {
	return nil
}

// attempt to coerce event as token transfer event.
func asTokenTransferEvent(gev *geEvent.Event) (*apievent.EventTokenTransfer, bool) {
	var err error
	var ok bool
	var ev apievent.EventTokenTransfer

	if gev.TxType != evTokenTransfer {
		return nil, false
	}

	pl := gev.Payload
	// we are assuming from and to are checksum addresses
	ev.From, ok = pl["from"].(string)
	if !ok {
		return nil, false
	}
	ev.To, ok = pl["to"].(string)
	if !ok {
		return nil, false
	}
	ev.TxHash, err = hex.NormalizeHex(gev.TxHash)
	if err != nil {
		logg.Errorf("could not decode tx hash", "tx", gev.TxHash, "err", err)
		return nil, false
	}

	value, ok := pl["value"].(string)
	if !ok {
		return nil, false
	}
	ev.Value, err = strconv.Atoi(value)
	if err != nil {
		logg.Errorf("could not decode value", "value", value, "err", err)
		return nil, false
	}

	ev.VoucherAddress = gev.ContractAddress

	return &ev, true
}
