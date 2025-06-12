package handlers

import (
	"fmt"
	"strings"
	"time"

	"meowBot/commands"
    "meowBot/util"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

const (
    StickerPackName = "whatsmeow"
    StickerAuthorName = "."
)

var startupTime = time.Now()
var authorizedUsers = map[string]bool{
        "6281391620354@s.whatsapp.net": true,
}

func isCreator(jid types.JID) bool {
        return authorizedUsers[jid.String()]
}

// HandleMessage dipanggil dari event_handler ketika ada pesan baru
func HandleMessage(v *events.Message) {
	// Ekstrak teks pesan
	var messageText string
	if v.Message.GetConversation() != "" {
		messageText = v.Message.GetConversation()
	} else if v.Message.GetExtendedTextMessage() != nil {
		messageText = v.Message.GetExtendedTextMessage().GetText()
	} else if v.Message.GetImageMessage() != nil {
        messageText = v.Message.GetImageMessage().GetCaption()
        } else if v.Message.GetVideoMessage() != nil {
        messageText = v.Message.GetVideoMessage().GetCaption()
    }


	log.Infof("Received message from %s: %s", v.Info.SourceString(), messageText)
	
	// Normalisasi teks perintah
	command := strings.ToLower(strings.TrimSpace(messageText))

        isStickerCmd := command == "!s" || command == "!tikel"
	hasMedia := v.Message.GetImageMessage() != nil || v.Message.GetVideoMessage() != nil ||
		(v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage() != nil &&
			(v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetImageMessage() != nil ||
				v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage().GetVideoMessage() != nil))

	switch {
	case command == "!ping":
		latency := time.Now().UnixMilli() % 1000
		replyText := fmt.Sprintf("*(づ｡◕‿‿◕｡づ) Pong! Latensi saat ini: %dms*", latency)
		util.SendReply(cli, v, replyText)

	case command == "!time":
		currentTime := time.Now().Format("15:04:05 - 02/01/2006")
		replyText := fmt.Sprintf("*⏰ Waktu server: %s*", currentTime)
		util.SendReply(cli, v, replyText)
	
	case command == "!uptime":
		uptime := time.Since(startupTime)
		replyText := fmt.Sprintf("*⏱️ Bot aktif selama: %v*", uptime.Round(time.Second))
		util.SendReply(cli, v, replyText)

        case isStickerCmd:
                if hasMedia {
			commands.HandleStickerCmd(cli, v, StickerPackName, StickerAuthorName)
		} else {
		replyText := "*🍭 Berikan gambar atau video*"
		util.SendReply(cli, v, replyText)
                }

	case strings.HasPrefix(messageText, "$ "):
                if !isCreator(v.Info.Sender) {
			return
		}
		
		cmdStr := strings.TrimPrefix(command, "$ ")
		
		parts := strings.Fields(cmdStr)
		command := parts[0]
		args := []string{}
		if len(parts) > 1 {
			args = parts[1:]
		}
		
		log.Infof("Executing command: %s %v", command, args)
		output, err := commands.ExecuteCommand(command, args)
		
		if err != nil {
			util.SendReply(cli, v, fmt.Sprintf("*🍓 Error:*\n```\n%s\n```", err.Error()))
			return
		}
		
		if len(output) > 4000 {
			output = output[:4000] + "\n\n... (output truncated)"
		}
		
		if strings.TrimSpace(output) == "" {
			return
		} else {
			util.SendReply(cli, v, fmt.Sprintf("*💻 Output:*\n```\n%s\n```", output))
		}
	}
}
