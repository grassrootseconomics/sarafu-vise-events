package event

import (
	"context"
	"fmt"

	geEvent "github.com/grassrootseconomics/eth-tracker/pkg/event"

	"git.defalsify.org/vise.git/logging"
	apievent "git.grassecon.net/grassrootseconomics/sarafu-api/event"
)

var (
	logg = logging.NewVanilla().WithDomain("term-event")
)

// Router is responsible for invoking handlers corresponding to events.
type Router struct {
	handler *apievent.EventsHandler
}

func NewRouter(handler *apievent.EventsHandler) *Router {
	return &Router{
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
	evCC, ok := asCustodialRegistrationEvent(gev)
	if ok {
		return r.handler.Handle(ctx, apievent.EventTokenTransferTag, evCC)
	}
	evTT, ok := asTokenTransferEvent(gev)
	if ok {
		return r.handler.Handle(ctx, apievent.EventRegistrationTag, evTT)
	}

	return fmt.Errorf("unexpected message")
}
