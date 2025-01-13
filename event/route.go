package event

import (
	"context"
	"fmt"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/sarafu-vise/store"
	viseevent "git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/event"
)

var (
	logg = logging.NewVanilla().WithDomain("term-event")
)

// Router is responsible for invoking handlers corresponding to events.
type Router struct {
	store storage.StorageService
	handler *viseevent.EventsHandler
}

func NewRouter(store storage.StorageService, handler *viseevent.EventsHandler) *Router {
	return &Router{
		store: store,
		handler: handler,
	}
}

// Route parses an event from the event stream, and resolves the handler
// corresponding to the event.
//
// An error will be returned if no handler can be found, or if the resolved
// handler fails to successfully execute.
func(r *Router) Route(ctx context.Context, gev *geEvent.Event) error {
	logg.DebugCtxf(ctx, "have event", "ev", gev)
	userDb, err := r.store.GetUserdataDb(ctx)
	if err != nil {
		return err
	}
	userStore := &store.UserDataStore{
		Db: userDb,
	}
	evCC, ok := asCustodialRegistrationEvent(gev)
	if ok {
		pr, err := r.store.GetPersister(ctx)
		if err != nil {
			return err
		}
		return r.handler.HandleCustodialRegistration(ctx, userStore, pr, evCC)
	}
	evTT, ok := asTokenTransferEvent(gev)
	if ok {
		return r.handler.HandleTokenTransfer(ctx, userStore, evTT)
	}

	return fmt.Errorf("unexpected message")
}
