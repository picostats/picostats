package main

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
)

var conf *config

var app *iris.Framework

var db *gorm.DB

func main() {
	initIris()

	initConfig()

	initDB()

	app.Get("/", home)

	app.Listen(":8080")
}
