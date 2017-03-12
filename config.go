package main

import (
	"encoding/json"
	"log"
	"os"
)

// Config represents a config object
type Config struct {
	DBUrl             string `json:"db_url"`
	AppUrl            string `json:"app_url"`
	RedisUrl          string `json:"redis_url"`
	LogSQL            bool   `json:"log_sql"`
	AllowRegistration bool   `json:"allow_registration"`
}

// Load method loads config file in config object
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

func initConfig() {
	conf = &Config{}
	conf.load("config.json")
}
