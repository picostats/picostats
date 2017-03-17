package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	DBType              string `json:"db_type"`
	DBUrl               string `json:"db_url"`
	AppUrl              string `json:"app_url"`
	RedisUrl            string `json:"redis_url"`
	ListenAddr          string `json:"listen_addr"`
	LogSQL              bool   `json:"log_sql"`
	Dev                 bool   `json:"dev"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	EncryptionKey       []byte `json:"encryption_key"`
	SparkpostKey        string `json:"sparkpost_key"`
	DemoLock            uint   `json:"demo_lock"`
	MaxWebsites         int    `json:"max_websites"`
}

func (c *Config) load(configFile string) error {
	file, err := os.Open(configFile)

	if err != nil {
		log.Printf("[config.go] Error while opening config file: %v", err)
		return err
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&c)

	if err != nil {
		log.Printf("[config.go] Error while decoding JSON: %v", err)
		return err
	}

	return nil
}

func initConfig() *Config {
	conf := &Config{}
	conf.load("config.json")
	return conf
}
