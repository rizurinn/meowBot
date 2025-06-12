package util

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var log waLog.Logger = waLog.Stdout("Utils", "INFO", true)

func SendReply(cli *whatsmeow.Client, v *events.Message, text string) {
	chatJID := v.Info.Chat

	ctxInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(v.Info.ID),
		Participant:   proto.String(v.Info.Sender.String()),
		QuotedMessage: v.Message,
	}

	replyMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: ctxInfo,
		},
	}

	_, err := cli.SendMessage(context.Background(), chatJID, replyMsg)
	if err != nil {
		log.Errorf("Gagal mengirim balasan: %v", err)
	}
}
