package redisstore

import (
	"github.com/go-redis/redis/v7"
	"gopkg.in/ini.v1"
	"time"
)

var client *redis.Client

func ConnectToRedis(config *ini.File) (*redis.Client, error) {
	client = redis.NewClient(&redis.Options{
		Addr:     config.Section("redis").Key("host").String(),
		Password: config.Section("redis").Key("password").String(),
		DB:       0,
	})

	_, err := client.Ping().Result()

	if err != nil {
		return client, err
	}

	return client, nil
}

func RedisSet(key string, value string) (bool, error) {
	err := client.Set(key, value, 1*time.Hour).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func RedisGet(key string) (string, error) {
	val, err := client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
