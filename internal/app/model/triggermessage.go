package model

type TriggerMessage struct {
	Id   int64  `pg:",pk"`
	Type string `pg:"type"`
	Text string `pg:"text"`
}
