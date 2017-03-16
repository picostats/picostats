package main

import (
	"log"

	sp "github.com/SparkPost/gosparkpost"
)

func sendTestMail() {
	apiKey := "59d81e5fe09c3d1480bb8b8610ac37a204d28740"
	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     apiKey,
		ApiVersion: 1,
	}
	var client sp.Client
	err := client.Init(cfg)
	if err != nil {
		log.Fatalf("SparkPost client init failed: %s\n", err)
	}

	tx := &sp.Transmission{
		Recipients: []string{"piha.tihomir@gmail.com"},
		Content: sp.Content{
			HTML:    "<p>Hello world</p>",
			From:    "PicoStats <no-reply@picostats.com>",
			Subject: "Hello from gosparkpost",
		},
	}
	id, _, err := client.Send(tx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Transmission sent with id [%s]\n", id)
}
