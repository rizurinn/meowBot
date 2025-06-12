package main

import (
	"context"
	"errors"
        "flag"
	"os"
	"os/signal"
	"syscall"
	
	"meowBot/handlers"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"

	"go.mau.fi/whatsmeow"
        waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	waBinary.IndentXML = true
	debugLogs := flag.Bool("debug", false, "Enable debug logs?")
	dbDialect := flag.String("db-dialect", "sqlite3", "Database dialect (sqlite3 or postgres)")
	dbAddress := flag.String("db-address", "file:session/whatsappbot.db?_foreign_keys=on", "Database address")
	usePairingCode := flag.Bool("pairing-code", false, "Use pairing code instead of QR code")
	phoneNumber := flag.String("phone", "", "Phone number for pairing code (format: +1234567890)")
	flag.Parse()

	logLevel := "INFO"
	if *debugLogs {
		logLevel = "DEBUG"
	}
	log := waLog.Stdout("Main", logLevel, true)

	dbLog := waLog.Stdout("Database", logLevel, true)
	storeContainer, err := sqlstore.New(context.Background(), *dbDialect, *dbAddress, dbLog)
	if err != nil {
		log.Errorf("Failed to connect to database: %v", err)
		return
	}
	device, err := storeContainer.GetFirstDevice(context.Background())
	if err != nil {
		log.Errorf("Failed to get device: %v", err)
		return
	}

	cli := whatsmeow.NewClient(device, waLog.Stdout("Client", logLevel, true))

	handlers.SetClient(cli)

	if cli.Store.ID == nil {
		if *usePairingCode && *phoneNumber != "" {
			code, err := cli.PairPhone(context.Background(), *phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
			if err != nil {
				log.Errorf("Failed to request pairing code: %v", err)
				return
			}
			log.Infof("Pairing code: %s", code)
		} else {
			ch, err := cli.GetQRChannel(context.Background())
			if err != nil {
				if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
					log.Errorf("Failed to get QR channel: %v", err)
				}
			} else {
				go func() {
					for evt := range ch {
						if evt.Event == "code" {
							qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
						} else {
							log.Infof("QR channel result: %s", evt.Event)
						}
					}
				}()
			}
		}
	}

	cli.AddEventHandler(handlers.MainEventHandler)

	err = cli.Connect()
	if err != nil {
		log.Errorf("Failed to connect: %v", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	log.Infof("Interrupt received, exiting")
	cli.Disconnect()
}
