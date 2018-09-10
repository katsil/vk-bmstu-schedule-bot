package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"github.com/BMSTU-bots/vk-bmstu-schedule-bot/config"
	"github.com/BMSTU-bots/vk-bmstu-schedule-bot/logger"
	vkapi "github.com/dimonchik0036/vk-api"
)

var (
	bot *vkapi.Client
)

func buildValidGroupID(groupID *string) (string, bool) {
	if ok, _ := regexp.Match(`^[а-яА-Я]{0,4}\d{0,2}\-\d{0,2}[а-яА-Я]?$`, []byte(*groupID)); !ok {
		return "", ok
	}
	return strings.ToUpper(*groupID), true
}

func parseSchedule(groupID *string) error {
	cmd := exec.Command("bmstu-schedule", *groupID, "-o", config.Instance.VaultPath)
	logger.Instance.Debug("Running parser",
		zap.String("GROUP", *groupID),
	)
	return cmd.Run()
}

func sendMessage(userID int64, text string) {
	bot.SendMessage(
		vkapi.NewMessage(
			vkapi.NewDstFromUserID(userID),
			text,
		),
	)
}

func sendNotFoundErrorMessage(userID int64, validGID string) {
	var msg string
	if unicode.IsNumber(rune(validGID[len(validGID)-1])) {
		msg = fmt.Sprintf(
			"Эээ, кажется, кто-то не уточнил "+
				"тип своей группы (Б/М/А). "+
				"Давай добавим соответствующую букву в "+
				"конце и попробуем еще раз. Например %sБ", validGID)
	} else {
		msg = fmt.Sprintf(
			"Чёт я ничего не нашел для группы %s. "+
				"Если проблема и правда во мне, то напиши @gabolaev", validGID)
	}
	sendMessage(userID, msg)
}

func msgHandler(fromID int64, text *string) {
	logger.Instance.Info("[NEW MESSAGE]",
		zap.Int64("FROM", fromID),
		zap.String("TEXT", *text),
	)

	validGID, ok := buildValidGroupID(text)
	if !ok {
		sendMessage(fromID, "Указан неверный формат номера группы.")
		return
	}

	sendMessage(fromID, "Пошел искать расписание для группы "+validGID)

	fileName := fmt.Sprintf("Расписание %s.ics", validGID)
	file, err := os.Open(config.Instance.VaultPath + fileName)
	if err != nil && parseSchedule(&validGID) != nil {
		sendNotFoundErrorMessage(fromID, validGID)
		return
	}

	file, err = os.Open(config.Instance.VaultPath + fileName)
	if err != nil {
		sendNotFoundErrorMessage(fromID, validGID)
		return
	}

	bot.SendDoc(
		vkapi.NewDstFromUserID(fromID),
		validGID+".ics",
		vkapi.FileReader{
			Reader: file,
			Size:   -1,
			Name:   file.Name(),
		},
	)
	sendMessage(fromID, "Тадам!")
	sendMessage(fromID, "Если вдруг будут проблемы при импорте "+
		"в календарь, можешь обращаться к @gabolaev")
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
