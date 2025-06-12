package commands

import (
	"context"
	"fmt"
	
	"meowBot/commands/api"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"meowBot/util"
)

var animLog waLog.Logger = waLog.Stdout("AnimStickerCmd", "INFO", true)

func HandleStickerCmd(cli *whatsmeow.Client, v *events.Message, packName, authorName string) {
	var media whatsmeow.DownloadableMessage
	var mediaType string

	if v.Message.GetImageMessage() != nil {
		media = v.Message.GetImageMessage()
		mediaType = "image"
	} else if v.Message.GetVideoMessage() != nil {
		media = v.Message.GetVideoMessage()
		mediaType = "video"
	} else if q := v.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage(); q != nil {
		if q.GetImageMessage() != nil {
			media = q.GetImageMessage()
			mediaType = "image"
		} else if q.GetVideoMessage() != nil {
			media = q.GetVideoMessage()
			mediaType = "video"
		}
	}

	mediaData, err := cli.Download(context.Background(), media)
	if err != nil {
		util.SendReply(cli, v, "*🍓 Gagal mengunduh media.*")
		return
	}

	stickerBytes, err := api.ConvertViaApi(mediaData, mediaType, packName, authorName)
	if err != nil {
		util.SendReply(cli, v, fmt.Sprintf("*🍓 Gagal membuat stiker:* %v", err))
		return
	}

	stickerMsg := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           nil,
			FileSHA256:    nil,
			FileEncSHA256: nil,
			MediaKey:      nil,
			Mimetype:      proto.String("image/webp"),
			Height:        proto.Uint32(512),
			Width:         proto.Uint32(512),
			DirectPath:    nil,
			FileLength:    proto.Uint64(uint64(len(stickerBytes))),
			IsAnimated:    proto.Bool(mediaType == "video"),
		},
	}
	
	uploaded, err := cli.Upload(context.Background(), stickerBytes, whatsmeow.MediaImage)
	if err != nil {
		util.SendReply(cli, v, "*🍓 Gagal mengunggah stiker.*")
		return
	}
	
	stickerMsg.StickerMessage.URL = proto.String(uploaded.URL)
	stickerMsg.StickerMessage.FileSHA256 = uploaded.FileSHA256
	stickerMsg.StickerMessage.FileEncSHA256 = uploaded.FileEncSHA256
	stickerMsg.StickerMessage.MediaKey = uploaded.MediaKey
	stickerMsg.StickerMessage.DirectPath = proto.String(uploaded.DirectPath)
	
	_, err = cli.SendMessage(context.Background(), v.Info.Chat, stickerMsg)
	if err != nil {
		animLog.Errorf("Gagal mengirim stiker: %v", err)
	}
}
