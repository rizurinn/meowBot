package handlers

import (
    "context"
    "encoding/json"
    "flag"
	"fmt"
    "os"
    "sync/atomic"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var cli *whatsmeow.Client
var log waLog.Logger

var logLevel = "INFO"
var historySyncID int32
var pushName = flag.String("push-name", "WhatsApp Bot", "Push name to set")

// SetClient dipanggil dari main.go untuk memberikan akses ke client WhatsApp
func SetClient(c *whatsmeow.Client) {
	cli = c
	log = waLog.Stdout("EventHandler", "INFO", true)
}

// MainEventHandler adalah handler utama yang merutekan event
func MainEventHandler(evt interface{}) {
switch v := evt.(type) {
	case *events.Message:
		HandleMessage(v)
	case *events.Receipt:
		if v.Type == events.ReceiptTypeRead || v.Type == events.ReceiptTypeReadSelf {
			log.Infof("%v was read by %s at %s", v.MessageIDs, v.SourceString(), v.Timestamp)
		} else if v.Type == events.ReceiptTypeDelivered {
			log.Infof("%s was delivered to %s at %s", v.MessageIDs[0], v.SourceString(), v.Timestamp)
		}
	case *events.Presence:
		if v.Unavailable {
			if v.LastSeen.IsZero() {
				log.Infof("%s is now offline", v.From)
			} else {
				log.Infof("%s is now offline (last seen: %s)", v.From, v.LastSeen)
			}
		} else {
			log.Infof("%s is now online", v.From)
		}
	case *events.HistorySync:
		id := atomic.AddInt32(&historySyncID, 1)
		fileName := fmt.Sprintf("session/history-%d-%d.json", startupTime, id)
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Errorf("Failed to open file to write history sync: %v", err)
			return
		}
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(v.Data)
		if err != nil {
			log.Errorf("Failed to write history sync: %v", err)
			return
		}
		log.Infof("Wrote history sync to %s", fileName)
		_ = file.Close()
	case *events.AppState:
		log.Debugf("App state event: %+v / %+v", v.Index, v.SyncActionValue)
	case *events.KeepAliveTimeout:
		log.Debugf("Keepalive timeout event: %+v", v)
		if v.ErrorCount > 3 {
			log.Debugf("Got >3 keepalive timeouts, forcing reconnect")
			go func() {
				cli.Disconnect()
				err := cli.Connect()
				if err != nil {
					log.Errorf("Failed to reconnect after keepalive timeout: %v", err)
				}
			}()
		}
	case *events.KeepAliveRestored:
		log.Debugf("Keepalive restored")
	case *events.LoggedOut:
		log.Infof("Logged out from WhatsApp: %v", v)
		if v.OnConnect {
			log.Infof("Logged out on connect, clearing store and reconnecting...")
			cli.RemoveEventHandlers()
			err := cli.Store.Delete(context.Background())
			if err != nil {
				log.Errorf("Failed to delete store: %v", err)
				return
			}
			cli = whatsmeow.NewClient(cli.Store, waLog.Stdout("Client", logLevel, true))
			cli.AddEventHandler(MainEventHandler)
			err = cli.Connect()
			if err != nil {
				log.Errorf("Failed to reconnect after logout: %v", err)
			}
		}
	case *events.Connected:
		log.Infof("Connected to WhatsApp")
		// if len(cli.Store.PushName) > 0 && cli.Store.PushName != *pushName {
		//	resp := cli.SendPresence(types.PresenceAvailable)
		//	log.Infof("Marked self as available (%s)", resp)
		// }
		// if cli.Store.PushName != *pushName {
		//	err := cli.SetStatusMessage(*pushName)
		//	if err != nil {
		//		log.Warnf("Failed to set push name: %v", err)
		//	}
		// }
	case *events.StreamReplaced:
		log.Infof("Stream replaced")
	case *events.ConnectFailure:
		log.Errorf("Connect failure: %v", v.Reason)
	case *events.ClientOutdated:
		log.Errorf("Client outdated")
	case *events.Disconnected:
		log.Infof("Disconnected from WhatsApp")
	}
}
