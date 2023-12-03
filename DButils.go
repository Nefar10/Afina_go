package main

import (
	"github.com/go-redis/redis"
)

func redisPing(client redis.Client) error {
	var err error
	_, err = client.Ping().Result()
	if err != nil {
		return err
	} else {
		return err
	}
}
