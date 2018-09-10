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

	if ok, _ := regexp.Match(`[а-яА-Я]{0,4}\d{0,2}\-\d{0,2}[а-яА-Я]?`, []byte(*groupID)); !ok {
		return "", ok
	}

	upperGID := strings.ToUpper(*groupID)
	if unicode.IsNumber(rune(upperGID[len(upperGID)-1])) {
		upperGID += "Б"
	}
	return upperGID, true
}

func parseSchedule(groupID *string) error {
	cmd := exec.Command("bmstu-schedule", *groupID, "-o", "vault")
	logger.Instance.Debug("Running parser",
		zap.String("GROUP", *groupID),
	)
	return cmd.Run()
}

func msgHandler(fromID int64, text *string) {
	logger.Instance.Info("[NEW MESSAGE]",
		zap.Int64("FROM", fromID),
		zap.String("TEXT", *text),
	)

	validGID, ok := buildValidGroupID(text)
	if !ok {
		bot.SendMessage(
			vkapi.NewMessage(
				vkapi.NewDstFromUserID(fromID),
				"Указан неверный формат номера группы.",
			),
		)
		return
	}

	fileName := fmt.Sprintf("Расписание %s.ics", validGID)

	file, err := os.Open("vault/" + fileName)
	if err != nil && parseSchedule(&validGID) != nil {
		groupCode := validGID[:len(validGID)-2]
		fmt.Println(groupCode)
		bot.SendMessage(
			vkapi.NewMessage(
				vkapi.NewDstFromUserID(fromID),
				fmt.Sprintf(`Расписание для группы %s не найдено. 
				Если вы не студент бакалавриата, укажите, пожалуйста, 
				тип группы путём добавления соответствующей буквы
				Например %sМ`, validGID, groupCode),
			),
		)
		return
	}

	file, err = os.Open("vault/" + fileName)
	if err != nil {
		logger.Instance.Error(err.Error())
		return
	}

	bot.SendDoc(
		vkapi.NewDstFromUserID(fromID),
		validGID,
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
