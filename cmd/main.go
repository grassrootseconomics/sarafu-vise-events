package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/config"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/event/nats"
	"git.grassecon.net/grassrootseconomics/sarafu-vise-events/lookup"
	viseevent "git.grassecon.net/grassrootseconomics/sarafu-vise/handlers/event"
)

var (
	logg          = logging.NewVanilla()
)

func main() {
	config.LoadConfig()

	var connStr string

	flag.StringVar(&connStr, "c", "", "connection string")
	flag.Parse()

	if connStr == "" {
		connStr = config.DbConn()
	}
	connData, err := storage.ToConnData(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connstr err: %v", err)
		os.Exit(1)
	}

	logg.Infof("start command", "conn", connData)

	ctx := context.Background()

	menuStorageService := storage.NewMenuStorageService(connData, "")

	eh := viseevent.NewEventsHandler(lookup.Api)
	n := nats.NewNatsSubscription(menuStorageService, eh)
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
