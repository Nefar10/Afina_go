package main

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func redisPing(client redis.Client) error {
	var err error
	SetCurOperation("Check DB connection", 1)
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
		SendToUser(gOwner, 0, gErr[10][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	} else {
		SendToUser(gOwner, 0, gIm[4][gLocale], MSG_INFO, 1, false)
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
		SendToUser(gOwner, 0, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}
	if len(keys) > 0 {
		for _, key := range keys {
			err = gRedisClient.Del(key).Err()
			if err != nil {
				SendToUser(gOwner, 0, gErr[10][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
			}
		}
	}
	keys, err = gRedisClient.Keys("ChatState:*").Result()
	if err != nil {
		SendToUser(gOwner, 0, gErr[12][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}
	for _, key := range keys {
		chatItem = GetChatStateDB(ParseChatKeyID(key))
		if chatItem.AllowState == CHAT_IN_PROCESS {
			DestroyChat(strconv.FormatInt(chatItem.ChatID, 10))
		}
	}
	SendToUser(gOwner, 0, "Кеш очищен.", MSG_INFO, 0, false)
}

func DBWrite(key string, value string, expiration time.Duration) error {
	SetCurOperation("Writting to DB", 0)
	var err = gRedisClient.Set(key, value, expiration).Err()
	if err != nil {
		SendToUser(gOwner, 0, gErr[10][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}
	return err
}

func DestroyChat(ID string) error {
	SetCurOperation("Clean DB", 0)
	err := gRedisClient.Del("ChatState:" + ID).Err()
	if err != nil {
		SendToUser(gOwner, 0, gErr[10][gLocale]+err.Error()+gIm[29][gLocale]+GetCurOperation(), MSG_ERROR, 0, false)
	}
	return err
}
