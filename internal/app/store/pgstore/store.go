package pgstore

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/vlladoff/trainingstatbot/internal/app/model"
	"gopkg.in/ini.v1"
)

var db *pg.DB

func ConnectToDb(config *ini.File) (*pg.DB, error) {
	db = pg.Connect(&pg.Options{
		Addr:     config.Section("postgresql").Key("host").String(),
		User:     config.Section("postgresql").Key("user").String(),
		Password: config.Section("postgresql").Key("password").String(),
		Database: config.Section("postgresql").Key("db").String(),
	})

	err := createSchemas(db)
	if err != nil {
		return db, err
	}
	err = appendDefaultValues(db)
	if err != nil {
		return db, err
	}

	return db, nil
}

func createSchemas(db *pg.DB) error {
	for _, model := range []interface{}{
		(*model.User)(nil),
		(*model.UserChat)(nil),
		(*model.Training)(nil),
		(*model.TrainingType)(nil),
		(*model.TriggerMessage)(nil),
	} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func appendDefaultValues(db *pg.DB) error {
	trainingType := new(model.TrainingType)
	dbErr := db.Model(trainingType).Limit(1).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return dbErr
	}

	if dbErr == pg.ErrNoRows {
		defaultTrainingTypes := []model.TrainingType{
			{
				Name: "Тренировка растяжки",
			},
			{
				Name: "Силовая тренировка",
			},
			{
				Name: "Кроссфит",
			},
			{
				Name: "Кардио тренировка",
			},
			{
				Name: "Отдых",
			},
		}
		_, err := db.Model(&defaultTrainingTypes).Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateTraining(userId int64, trainingTypeId, date int) (bool, error) {
	training := model.Training{
		UserId:         userId,
		TrainingTypeId: trainingTypeId,
		Date:           date,
	}
	_, err := db.Model(&training).Insert()
	if err != nil {
		return false, err
	}
	return true, nil
}

func SelectTrainingTypes() ([]model.TrainingType, error) {
	var trainingTypes []model.TrainingType
	dbErr := db.Model(&trainingTypes).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return trainingTypes, nil
}

func SelectMessages() ([]model.TriggerMessage, error) {
	var messages []model.TriggerMessage
	dbErr := db.Model(&messages).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return messages, nil
}

func SelectUser(tgUserId int) (model.User, error) {
	user := new(model.User)
	dbErr := db.Model(user).Where("tg_user_id = ?", tgUserId).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return *user, dbErr
	}
	return *user, nil
}

func CreateUser(user model.User) (model.User, error) {
	_, dbErr := db.Model(&user).Returning("*").Insert()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return user, dbErr
	}
	return user, nil
}

func SelectUsers(userIds []int64) ([]model.User, error) {
	var users []model.User
	dbErr := db.Model(&users).Where("id in (?)", pg.In(userIds)).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return users, nil
}

func UpdateLastUserChatId(userId, chatId int64) (bool, error) {
	var user model.User
	_, err := db.Model(&user).Set("last_chat_id = ?", chatId).Where("id = ?", userId).Update()
	if err != nil {
		return false, err
	}

	return true, nil
}

func SelectUserChats(chatId int64) ([]model.UserChat, error) {
	var userChats []model.UserChat
	dbErr := db.Model(&userChats).Where("chat_id = ?", chatId).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return userChats, nil
}

func SelectTrainingByUser(userId int64) ([]model.Training, error) {
	var trainings []model.Training
	dbErr := db.Model(&trainings).Where("user_id = ?", userId).Select()
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return trainings, nil
}

func SelectUserChatsByUserId(userId int64) ([]int64, error) {
	var userChats []int64
	dbErr := db.Model((*model.UserChat)(nil)).Column("chat_id").Where("user_id = ?", userId).Select(&userChats)
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}
	return userChats, nil
}

func CreateUserChat(userId, chatId int64) (bool, error) {
	userChat := model.UserChat{
		UserId: userId,
		ChatId: chatId,
	}
	_, err := db.Model(&userChat).Insert()
	if err != nil {
		return false, err
	}

	return true, nil
}

func CountTrainingByUsers(userIds []int64) ([]model.TopTrainingUsers, error) {
	var topTrainingUsers []model.TopTrainingUsers
	dbErr := db.Model((*model.Training)(nil)).
		Column("user_id").
		ColumnExpr("count(*) AS training_count").
		Where("user_id in (?)", pg.In(userIds)).
		Group("user_id").
		Order("training_count DESC").
		Select(&topTrainingUsers)
	if dbErr != nil && dbErr != pg.ErrNoRows {
		return nil, dbErr
	}

	return topTrainingUsers, nil
}
