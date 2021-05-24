package model

type User struct {
	Id         int64  `pg:",pk"`
	TgUserId   int    `pg:"tg_user_id,nopk"`
	LastChatId int64  `pg:"last_chat_id,nopk"`
	FirstName  string `pg:"first_name"`
	LastName   string `pg:"last_name"`
	Username   string `pg:"username"`
	Alias      string `pg:"alias,unique"`
}

type UserChat struct {
	UserId int64 `pg:"user_id,nopk"`
	ChatId int64 `pg:"chat_id,nopk"`
}

type TopTrainingUsers struct {
	UserId        int64
	TrainingCount int
}
