package lookup

import (
	"context"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	storedb "git.grassecon.net/grassrootseconomics/sarafu-vise/store/db"
	"git.grassecon.net/grassrootseconomics/common/hex"
)

var (
	logg = logging.NewVanilla().WithDomain("term-lookup")
)

// Identity contains all flavors of identifiers used across stream, api and
// client for a single agent.
type Identity struct {
	NormalAddress string
	ChecksumAddress string
	SessionId string
}

// IdentityFromAddress fully populates and Identity object from a given
// checksum address.
//
// It is the caller's responsibility to ensure that a valid checksum address
// is passed.
func IdentityFromAddress(ctx context.Context, userStore *store.UserDataStore, address string) (Identity, error) {
	var err error
	var ident Identity

	ident.ChecksumAddress = address
	ident.NormalAddress, err = hex.NormalizeHex(ident.ChecksumAddress)
	if err != nil {
		return ident, err
	}
	ident.SessionId, err = getSessionIdByAddress(ctx, userStore, ident.NormalAddress)
	if err != nil {
		return ident, err
	}
	return ident, nil
}

// load matching session from address from db store.
func getSessionIdByAddress(ctx context.Context, userStore *store.UserDataStore, address string) (string, error) {
	// TODO: replace with userdatastore when double sessionid issue fixed
	//r, err := store.ReadEntry(ctx, address, common.DATA_PUBLIC_KEY_REVERSE)
	userStore.Db.SetPrefix(db.DATATYPE_USERDATA)
	userStore.Db.SetSession(address)
	r, err := userStore.Db.Get(ctx, storedb.PackKey(storedb.DATA_PUBLIC_KEY_REVERSE, []byte{}))
	if err != nil {
		return "", err
	}
	return string(r), nil
}
