package event

import (
	"context"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
	"git.grassecon.net/grassrootseconomics/visedriver/testutil/mocks"
	"git.grassecon.net/grassrootseconomics/sarafu-api/remote/http"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/config"
	"git.grassecon.net/grassrootseconomics/common/hex"
	storedb "git.grassecon.net/grassrootseconomics/sarafu-vise/store/db"
	apievent "git.grassecon.net/grassrootseconomics/sarafu-api/event"
	apimocks "git.grassecon.net/grassrootseconomics/sarafu-api/testutil/mocks"
	viseevent "git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/event"
)


func TestCustodialRegistration(t *testing.T) {
	err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	storageService := mocks.NewMemStorageService(ctx)
	userDb := storageService.Db

	alice, err := hex.NormalizeHex(apimocks.AliceChecksum)
	if err != nil {
		t.Fatal(err)
	}

	userDb.SetSession(alice)
	userDb.SetPrefix(db.DATATYPE_USERDATA)
	err = userDb.Put(ctx, storedb.PackKey(storedb.DATA_PUBLIC_KEY_REVERSE, []byte{}), []byte(apimocks.AliceSession))
	if err != nil {
		t.Fatal(err)
	}

	st := state.NewState(248)
	ca := cache.NewCache()
	pr, _ := storageService.GetPersister(ctx)
	pr = pr.WithContent(st, ca)
	err = pr.Save(apimocks.AliceSession)
	if err != nil {
		t.Fatal(err)
	}

	ev := &apievent.EventCustodialRegistration{
		Account: apimocks.AliceChecksum,
	}

	// Use dev service or mock service instead
	eu := viseevent.NewEventsUpdater(&http.HTTPAccountService{}, storageService)
	eh := eu.ToEventsHandler()
	err = eh.Handle(ctx, apievent.EventRegistrationTag, ev)
	if err != nil {
		t.Fatal(err)
	}
}
