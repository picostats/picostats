package main

import (
	"log"

	"gopkg.in/redis.v5"
)

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.RedisUrl,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("[redis.go] Error: %s", err)
	}

	return client
}
