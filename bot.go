package main

import (
	"log"

	config "github.com/BMSTU-bots/vk-bmstu-schedule-bot/config"
	vkapi "github.com/dimonchik0036/vk-api"
)

var (
	bot *vkapi.Client
)

func msgHandler(fromID int64, text string) {
	log.Printf("Handling new message from %d: %s", fromID, text)
	bot.SendMessage(
		vkapi.NewMessage(vkapi.NewDstFromUserID(fromID),
			"👌",
		))
}

func main() {
	var err error

	if bot, err = vkapi.NewClientFromToken(config.Instance.AccessToken); err != nil {
		log.Panic(err)
	}

	bot.Log(true)

	if err := bot.InitLongPoll(0, 2); err != nil {
		log.Panic(err)
	}

	LPCfg := vkapi.LPConfig{
		Wait: 25,
		Mode: vkapi.LPModeAttachments,
	}
	updates, _, err := bot.GetLPUpdatesChan(100, LPCfg)
	if err != nil {
		log.Panic(err)
	}

	var msg *vkapi.LPMessage

	for update := range updates {
		msg = update.Message

		if msg == nil || !update.IsNewMessage() || msg.Outbox() {
			continue
		}
		go msgHandler(msg.FromID, msg.Text)
	}
}
