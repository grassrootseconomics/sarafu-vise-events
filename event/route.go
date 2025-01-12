package event

import (
	"context"
	"fmt"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
)

var (
	logg = logging.NewVanilla().WithDomain("term-event")
)

// Router is responsible for invoking handlers corresponding to events.
type Router struct {
	Store storage.StorageService
}

// Route parses an event from the event stream, and resolves the handler
// corresponding to the event.
//
// An error will be returned if no handler can be found, or if the resolved
// handler fails to successfully execute.
func(r *Router) Route(ctx context.Context, gev *geEvent.Event) error {
	logg.DebugCtxf(ctx, "have event", "ev", gev)
	userDb, err := r.Store.GetUserdataDb(ctx)
	if err != nil {
		return err
	}
	userStore := &store.UserDataStore{
		Db: userDb,
	}
	evCC, ok := asCustodialRegistrationEvent(gev)
	if ok {
		pr, err := r.Store.GetPersister(ctx)
		if err != nil {
			return err
		}
		return handleCustodialRegistration(ctx, userStore, pr, evCC)
	}
	evTT, ok := asTokenTransferEvent(gev)
	if ok {
		return handleTokenTransfer(ctx, userStore, evTT)
	}

	return fmt.Errorf("unexpected message")
}
