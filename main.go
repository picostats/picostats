package main

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
)

var app *iris.Framework

var conf *Config

var db *gorm.DB

func main() {
	// Loads and parses config.json file to struct
	conf = initConfig()

	// Initializes Iris web framework
	app = initIris()

	// Connects to the database and does automatic migrations
	db = initDB()

	// Initizalizes session and session cookie
	initSession()

	// GET view handlers
	app.Get(APP_PATH, loginRequired, dashboardView)
	app.Get(APP_PATH+"/sign-in", signInView)
	app.Get(APP_PATH+"/sign-up", signUpView)

	// POST view handlers
	app.Post(APP_PATH+"/sign-in", signInPostView)
	app.Post(APP_PATH+"/sign-up", signUpPostView)

	app.Listen(conf.ListenAddr)
}
