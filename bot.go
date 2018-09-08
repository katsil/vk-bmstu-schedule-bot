package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/BMSTU-bots/vk-bmstu-schedule-bot/config"
	"github.com/BMSTU-bots/vk-bmstu-schedule-bot/logger"
	vkapi "github.com/dimonchik0036/vk-api"
)

var (
	bot *vkapi.Client
)

func msgHandler(fromID int64, text *string) {
	logger.Instance.Info("[NEW MESSAGE]",
		zap.Int64("FROM", fromID),
		zap.String("TEXT", *text),
	)
	file, err := os.Open("vault/schedule.ics")
	if err != nil {
		logger.Instance.Error(err.Error())
	}

	bot.SendDoc(
		vkapi.NewDstFromUserID(fromID),
		"schedule.ics",
		vkapi.FileReader{
			Reader: file,
			Size:   -1,
			Name:   file.Name(),
		},
	)
}

func main() {
	var err error

	if bot, err = vkapi.NewClientFromToken(config.Instance.AccessToken); err != nil {
		logger.Instance.Error(err.Error())
	}

	bot.Log(config.Instance.BotLog)

	if err := bot.InitLongPoll(0, 2); err != nil {
		logger.Instance.Error(err.Error())
	}

	LPCfg := vkapi.LPConfig{
		Wait: 25,
		Mode: vkapi.LPModeAttachments,
	}
	updates, _, err := bot.GetLPUpdatesChan(100, LPCfg)
	if err != nil {
		logger.Instance.Error(err.Error())
	}

	var msg *vkapi.LPMessage

	for update := range updates {
		msg = update.Message

		if msg == nil || !update.IsNewMessage() || msg.Outbox() {
			continue
		}
		go msgHandler(msg.FromID, &msg.Text)
	}
}
