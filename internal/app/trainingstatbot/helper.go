package trainingstatbot

import (
	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vlladoff/trainingstatbot/internal/app/model"
	"github.com/vlladoff/trainingstatbot/internal/app/store/pgstore"
	"github.com/vlladoff/trainingstatbot/internal/app/store/redisstore"
	"log"
	"strconv"
	"strings"
)

func AddTraining(userId int64, trainingTypeId, date int) bool {
	status, err := pgstore.CreateTraining(userId, trainingTypeId, date)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return status
}

func GetUser(userId int) model.User {
	user, err := pgstore.SelectUser(userId)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return user
}

func RegisterUser(user model.User) model.User {
	createdUser, err := pgstore.CreateUser(user)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	AddUserChat(createdUser.Id, createdUser.LastChatId, 0)

	return createdUser
}

func AddUserChat(userId, currentChatId, lastChatId int64) bool {
	if currentChatId == lastChatId {
		return true
	}

	_, err := pgstore.UpdateLastUserChatId(userId, currentChatId)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	userChats, err := pgstore.SelectUserChatsByUserId(userId)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	for _, chatId := range userChats {
		if chatId == currentChatId {
			return true
		}
	}

	status, err := pgstore.CreateUserChat(userId, currentChatId)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return status
}

func SetLastUserAction(userId int64, actionName string) bool {
	status, err := redisstore.RedisSet("user_"+strconv.FormatInt(userId, 10)+"_last_action", actionName)
	if err != nil {
		log.Printf("Redis error: %s", err)
	}

	return status
}

func GetUserLastAction(userId int64) string {
	result, err := redisstore.RedisGet("user_" + strconv.FormatInt(userId, 10) + "_last_action")
	if err != nil && err != redis.Nil {
		log.Printf("Redis error: %s", err)

		return ""
	}

	return result
}

func PrepareUser(update tgbotapi.Update) model.User {
	user := model.User{
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
		Username:   update.Message.From.UserName,
		TgUserId:   update.Message.From.ID,
		LastChatId: update.Message.Chat.ID,
	}

	return user
}

func MessageSearch(str string, slc []string) bool {
	for _, val := range slc {
		if strings.ToLower(val) == strings.ToLower(str) {
			return true
		}
	}

	return false
}

func GetStatistics(userId int64) string {
	trainings, err := pgstore.SelectTrainingByUser(userId)
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	countTrainings := make(map[int]int, 10)
	for _, val := range trainings {
		countTrainings[val.TrainingTypeId]++
	}

	responseMessage := "Всего тренировок: " + strconv.Itoa(len(trainings)) + "\n\n"
	for i := 1; i <= len(TrainingTypes); i++ {
		if countTrainings[i] == 0 {
			continue
		}
		responseMessage += TrainingTypes[i] + ": " + strconv.Itoa(countTrainings[i]) + "\n"
	}

	return responseMessage
}

func GetTop(chatId int64) string {
	currentChatUsers, err := pgstore.SelectUserChats(chatId)
	if err != nil {
		log.Printf("DB error: %s", err)
		return ""
	}

	var userIds []int64
	for _, user := range currentChatUsers {
		userIds = append(userIds, user.UserId)
	}

	trainings, err := pgstore.CountTrainingByUsers(userIds)
	if err != nil {
		log.Printf("DB error: %s", err)
		return ""
	}

	var trainingsUsers []int64
	for _, training := range trainings {
		trainingsUsers = append(trainingsUsers, training.UserId)
	}

	users, err := pgstore.SelectUsers(trainingsUsers)
	if err != nil {
		log.Printf("DB error: %s", err)
		return ""
	}

	statsUsers := make(map[int64]model.User)
	for _, user := range users {
		statsUsers[user.Id] = user
	}

	responseMessage := "Топ атлетов:\n\n"
	for i, training := range trainings {
		responseMessage += strconv.Itoa(i+1) + ".🤸" + statsUsers[training.UserId].FirstName + ": " +
			strconv.Itoa(training.TrainingCount) + "\n"
	}

	return responseMessage
}
