package event

import (
	"context"
	"testing"

	memdb "git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/common"
	"git.grassecon.net/term/internal/testutil"
)

func TestCustodialRegistration(t *testing.T) {
	err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	userDb := memdb.NewMemDb()
	err = userDb.Connect(ctx, "")
	if err != nil {
		panic(err)
	}

	alice, err := common.NormalizeHex(testutil.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, common.PackKey(common.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(testutil.AliceSession))
	if err != nil {
		t.Fatal(err)
	}
	store := common.UserDataStore{
		Db: userDb,
	}

	st := state.NewState(248)
	ca := cache.NewCache()
	pr := persist.NewPersister(userDb)
	pr = pr.WithContent(st, ca)
	err = pr.Save(testutil.AliceSession)
	if err != nil {
		t.Fatal(err)
	}

	ev := &eventCustodialRegistration{
		Account: testutil.AliceChecksum,
	}
	err = handleCustodialRegistration(ctx, &store, pr, ev)
	if err != nil {
		t.Fatal(err)
	}

}
