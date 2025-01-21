package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/config"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/event/nats"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
	viseevent "git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/event"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
)

var (
	logg = logging.NewVanilla()
)

func main() {
	config.LoadConfig()

	override := config.NewOverride()

	flag.StringVar(override.DbConn, "c", "?", "default connection string (replaces all unspecified strings)")
	flag.StringVar(override.ResourceConn, "resource", "?", "resource connection string")
	flag.StringVar(override.UserConn, "userdata", "?", "userdata store connection string")
	flag.StringVar(override.StateConn, "state", "?", "state store connection string")
	flag.Parse()

	config.Apply(&override)
	conns, err := config.GetConns()
	if err != nil {
		fmt.Fprintf(os.Stderr, "conn specification error: %v\n", err)
		os.Exit(1)
	}

	logg.Infof("start command", "conn", conns)

	ctx := context.Background()

	menuStorageService := storage.NewMenuStorageService(conns)

	eu := viseevent.NewEventsUpdater(lookup.Api, menuStorageService)
	eh := eu.ToEventsHandler()
	n := nats.NewNatsSubscription(eh)
	err = n.Connect(ctx, config.JetstreamURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Stream connect err: %v", err)
		os.Exit(1)
	}
	defer n.Close()

	cint := make(chan os.Signal)
	cterm := make(chan os.Signal)
	signal.Notify(cint, os.Interrupt, syscall.SIGINT)
	signal.Notify(cterm, os.Interrupt, syscall.SIGTERM)
	select {
	case _ = <-cint:
	case _ = <-cterm:
	}
}
