package event

import (
	"context"
	"fmt"
	"strings"
	"strconv"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"

	"git.defalsify.org/vise.git/db"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	"git.grassecon.net/grassrootseconomics/common/hex"
	storedb "git.grassecon.net/grassrootseconomics/sarafu-vise/store/db"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
)

const (
	evTokenTransfer = "TOKEN_TRANSFER"
	// TODO: export from visedriver storage package
	DATATYPE_USERSUB = 64
)

// fields used for handling token transfer event.
type eventTokenTransfer struct {
	To string
	Value int
	VoucherAddress string
	TxHash string
	From string
}

type eventTokenMint struct {
	To string
	Value int
	TxHash string
	VoucherAddress string
}


// formatter for transaction data
//
// TODO: current formatting is a placeholder.
func formatTransaction(idx int, tx dataserviceapi.Last10TxResponse) string {
	return fmt.Sprintf("%d %s %s", idx, tx.DateBlock, tx.TxHash[:10])
}

// refresh and store transaction history.
func updateTokenTransferList(ctx context.Context, userStore *store.UserDataStore, identity lookup.Identity) error {
	var r []string

	txs, err := lookup.Api.FetchTransactions(ctx, identity.ChecksumAddress)
	if err != nil {
		return err
	}

	for i, tx := range(txs) {
		r = append(r, formatTransaction(i, tx))
	}

	s := strings.Join(r, "\n")

	return userStore.WriteEntry(ctx, identity.SessionId, storedb.DATA_TRANSACTIONS, []byte(s))
}

// refresh and store token list.
func updateTokenList(ctx context.Context, userStore *store.UserDataStore, identity lookup.Identity) error {
	holdings, err := lookup.Api.FetchVouchers(ctx, identity.ChecksumAddress)
	if err != nil {
		return err
	}
	metadata := store.ProcessVouchers(holdings)
	_ = metadata

	// TODO: make sure subprefixdb is thread safe when using gdbm
	// TODO: why is address session here unless explicitly set
	userStore.Db.SetSession(identity.SessionId)
	userStore.Db.SetPrefix(DATATYPE_USERSUB)

	k := append([]byte("vouchers"), []byte("sym")...)
	err = userStore.Db.Put(ctx, k, []byte(metadata.Symbols))
	if err != nil {
		return err
	}
	logg.TraceCtxf(ctx, "processvoucher", "key", k)
	k = append([]byte("vouchers"), []byte("bal")...)
	err = userStore.Db.Put(ctx, k, []byte(metadata.Balances))
	if err != nil {
		return err
	}

	k = append([]byte("vouchers"), []byte("deci")...)
	err = userStore.Db.Put(ctx, k, []byte(metadata.Decimals))
	if err != nil {
		return err
	}

	k = append([]byte("vouchers"), []byte("addr")...)
	err = userStore.Db.Put(ctx, k, []byte(metadata.Addresses))
	if err != nil {
		return err
	}

	return nil
}

// set default token to given symbol.
func updateDefaultToken(ctx context.Context, userStore *store.UserDataStore, identity lookup.Identity, activeSym string) error {
	pfxDb := store.StoreToPrefixDb(userStore, []byte("vouchers"))
	// TODO: the activeSym input should instead be newline separated list?
	tokenData, err := store.GetVoucherData(ctx, pfxDb, activeSym)
	if err != nil {
		return err
	}
	return store.UpdateVoucherData(ctx, userStore, identity.SessionId, tokenData)
}

// waiter to check whether object is available on dependency endpoints.
func updateWait(ctx context.Context) error {
	return nil
}

// use api to resolve address to token symbol.
func toSym(ctx context.Context, address string) ([]byte, error) {
	voucherData, err := lookup.Api.VoucherData(ctx, address)
	if err != nil {
		return nil, err
	}
	return []byte(voucherData.TokenSymbol), nil
}

// execute all 
func updateToken(ctx context.Context, userStore *store.UserDataStore, identity lookup.Identity, tokenAddress string) error {
	err := updateTokenList(ctx, userStore, identity)
	if err != nil {
		return err
	}

	userStore.Db.SetSession(identity.SessionId)
	activeSym, err := userStore.ReadEntry(ctx, identity.SessionId, storedb.DATA_ACTIVE_SYM)
	if err == nil {
		return nil
	}
	if !db.IsNotFound(err) {
		return err
	}
	if activeSym == nil {
		activeSym, err = toSym(ctx, tokenAddress)
		if err != nil {
			return err
		}
	}
	logg.Debugf("barfoo")

	err = updateDefaultToken(ctx, userStore, identity, string(activeSym))
	if err != nil {
		return err
	}

	err = updateTokenTransferList(ctx, userStore, identity)
	if err != nil {
		return err
	}

	return nil
}

// attempt to coerce event as token transfer event.
func asTokenTransferEvent(gev *geEvent.Event) (*eventTokenTransfer, bool) {
	var err error
	var ok bool
	var ev eventTokenTransfer

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

// handle token transfer.
//
// if from and to are NOT the same, handle code will be executed once for each side of the transfer.
func handleTokenTransfer(ctx context.Context, userStore *store.UserDataStore, ev *eventTokenTransfer) error {
	identity, err := lookup.IdentityFromAddress(ctx, userStore, ev.From)
	if err != nil {
		if !db.IsNotFound(err) {
			return err
		}
	} else {
		err = updateToken(ctx, userStore, identity, ev.VoucherAddress)
		if err != nil {
			return err
		}
	}

	if strings.Compare(ev.To, ev.From) != 0 {
		identity, err = lookup.IdentityFromAddress(ctx, userStore, ev.To)
		if err != nil {
			if !db.IsNotFound(err) {
				return err
			}
		} else {
			err = updateToken(ctx, userStore, identity, ev.VoucherAddress)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// handle token mint.
func handleTokenMint(ctx context.Context, userStore *store.UserDataStore, ev *eventTokenMint) error {
	identity, err := lookup.IdentityFromAddress(ctx, userStore, ev.To)
	if err != nil {
		if !db.IsNotFound(err) {
			return err
		}
	} else {
		err = updateToken(ctx, userStore, identity, ev.VoucherAddress)
		if err != nil {
			return err
		}
	}
	return nil
}
