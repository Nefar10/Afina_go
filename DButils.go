package main

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func redisPing(client redis.Client) error {
	var err error
	gCurProcName = "Check DB connection"
	_, err = client.Ping().Result()
	if err != nil {
		return err
	} else {
		return err
	}
}

func ResetDB() {
	var err error
	SetCurOperation("Resetting", 0)
	err = gRedisClient.FlushDB().Err()
	if err != nil {
		SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	} else {
		SendToUser(gOwner, IM4[gLocale], INFO, 1)
		os.Exit(0)
	}
}

func FlushCache() {
	var keys []string
	var err error
	var chatItem ChatState
	SetCurOperation("Cache cleaning", 0)
	keys, err = gRedisClient.Keys("QuestState:*").Result()
	if err != nil {
		SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	if len(keys) > 0 {
		for _, key := range keys {
			err = gRedisClient.Del(key).Err()
			if err != nil {
				SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
			}
		}
	}
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, E12[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	for _, key := range keys {
		chatItem = GetChatStateDB(key)
		if chatItem.AllowState == IN_PROCESS {
			DestroyChat(strconv.FormatInt(chatItem.ChatID, 10))
		}
	}
	SendToUser(gOwner, "Кеш очищен.", INFO, 0)
}

func DBWrite(key string, value string, expiration time.Duration) error {
	var err = gRedisClient.Set(key, value, expiration).Err()
	if err != nil {
		SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	return err
}

func DestroyChat(ID string) error {
	err := gRedisClient.Del("ChatState:" + ID).Err()
	if err != nil {
		SendToUser(gOwner, E10[gLocale]+err.Error()+IM29[gLocale]+gCurProcName, ERROR, 0)
	}
	return err
}
