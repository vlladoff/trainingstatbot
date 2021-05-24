package model

type Training struct {
	Id             int64 `pg:",pk"`
	UserId         int64 `pg:"user_id,nopk"`
	TrainingTypeId int   `pg:"training_type_id,nopk"`
	Date           int   `pg:"date,nopk"`
}

type TrainingType struct {
	Id   int    `pg:",pk"`
	Name string `pg:"name"`
}
