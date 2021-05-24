package trainingstatbot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vlladoff/trainingstatbot/internal/app/model"
	"github.com/vlladoff/trainingstatbot/internal/app/store/pgstore"
	"github.com/vlladoff/trainingstatbot/internal/app/store/redisstore"
	"gopkg.in/ini.v1"
	"log"
	"math/rand"
	"strconv"
)

var (
	Config      *ini.File
	Bot         *tgbotapi.BotAPI
	CurrentUser model.User

	//messages
	TrainingStart []string
	TrainingEnd   []string
	TrainingStats []string
	TrainingTop   []string

	TrainingTypes map[int]string
)

func Start() error {
	if cfg, err := ini.Load("configs/trainingstatbot.ini"); err != nil {
		log.Printf("Fail to read file: %v", err)
		return err
	} else {
		Config = cfg
	}

	db, err := pgstore.ConnectToDb(Config)
	if err != nil {
		return err
	}
	defer db.Close()

	redis, err := redisstore.ConnectToRedis(Config)
	if err != nil {
		return err
	}
	defer redis.Close()

	LoadDefaultData()

	Bot, err = tgbotapi.NewBotAPI(Config.Section("main").Key("tg_token").String())
	if err != nil {
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Printf("Authorized on account %s", Bot.Self.UserName)

	updates, _ := Bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		ProcessMessage(update)
	}

	return nil
}

func LoadDefaultData() {
	defaultMessages, err := pgstore.SelectMessages()
	if err != nil {
		log.Printf("Db error: %s", err)
	}

	for _, val := range defaultMessages {
		if val.Type == "training_start" {
			TrainingStart = append(TrainingStart, val.Text)
		}
		if val.Type == "training_end" {
			TrainingEnd = append(TrainingEnd, val.Text)
		}
		if val.Type == "training_stats" {
			TrainingStats = append(TrainingStats, val.Text)
		}
		if val.Type == "training_top" {
			TrainingTop = append(TrainingTop, val.Text)
		}
	}

	trainingTypes, err := pgstore.SelectTrainingTypes()
	if err != nil {
		log.Printf("Db error: %s", err)
	}

	TrainingTypes = make(map[int]string)
	for _, val := range trainingTypes {
		TrainingTypes[val.Id] = val.Name
	}
}

func ProcessMessage(update tgbotapi.Update) {
	CurrentUser = PrepareUser(update)
	if user := GetUser(CurrentUser.TgUserId); user.Id == 0 {
		CurrentUser = RegisterUser(CurrentUser)
	} else {
		CurrentUser = user
	}
	AddUserChat(CurrentUser.Id, update.Message.Chat.ID, CurrentUser.LastChatId)
	lastUserAction := GetUserLastAction(CurrentUser.Id)

	//Training start
	if MessageSearch(update.Message.Text, TrainingStart) {
		tgMsg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, "assets/images/training_types.png")
		tgMsg.ParseMode = "markdown"
		tgMsg.ReplyToMessageID = update.Message.MessageID
		tgMsg.Caption = TrainingEnd[rand.Intn(len(TrainingEnd))] + "\n\nА как вы потренировались? 💪\nВведите тип тренировки: "
		Bot.Send(tgMsg)

		SetLastUserAction(CurrentUser.Id, "trainingStart")
		return
	}

	//Training end
	if lastUserAction == "trainingStart" {
		if trainingTypeId, _ := strconv.Atoi(update.Message.Text); trainingTypeId >= 1 && trainingTypeId <= 10 {
			if AddTraining(CurrentUser.Id, trainingTypeId, update.Message.Date) {
				Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Запомнил! Жду когда потренируетесь снова 🏋️"))
				SetLastUserAction(CurrentUser.Id, "")

				return
			}
		} else {
			Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"Неверный формат ввода! Необходимо ввести число от 1 до "+strconv.Itoa(len(TrainingTypes))))

			return
		}
	}

	//Training stats
	if MessageSearch(update.Message.Text, TrainingStats) {
		Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, GetStatistics(CurrentUser.Id)))

		return
	}

	//Training top from chat
	if MessageSearch(update.Message.Text, TrainingTop) {
		Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, GetTop(update.Message.Chat.ID)))

		return
	}
}
