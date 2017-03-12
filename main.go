package main

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
)

var app *iris.Framework

var conf *Config

var db *gorm.DB

func main() {
	// Initializes Iris web framework
	initIris()

	// Loads and parses config.json file to struct
	initConfig()

	// Connects to the database and does automatic migrations
	initDB()

	// GET view handlers
	app.Get("/", home)

	// POST view handlers

	app.Listen(":8080")
}
