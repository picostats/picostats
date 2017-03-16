package main

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/redis.v5"
)

var app *iris.Framework

var conf *Config

var db *gorm.DB

var red *redis.Client

var clip *CliParser

func main() {
	// Loads and parses config.json file to struct
	conf = initConfig()

	// Parses CLI options and arguments
	initCli()

	// Initializes Iris web framework
	app = initIris()

	// Connects to the database and does automatic migrations
	db = initDB()

	// Initizalizes session and session cookie
	initSession()

	// Initializes Redis connection
	red = initRedis()

	// Initializes worker and starts saving data
	initWorker()

	// GET view handlers
	app.Get(APP_PATH+"/sign-in", signInView)
	app.Get(APP_PATH+"/sign-up", signUpView)
	app.Get(APP_PATH+"/sign-out", signOutView)
	app.Get(APP_PATH+"/account", loginRequired, accountView)
	app.Get(APP_PATH+"/account/delete", loginRequired, accountDeleteView)
	app.Get(APP_PATH+"/websites/new", loginRequired, newWebsiteView)
	app.Get(APP_PATH+"/websites/delete/{id}", loginRequired, websiteDeleteView)
	app.Get(APP_PATH+"/websites/default/{id}", loginRequired, websiteMakeDefaultView)
	app.Get(APP_PATH+"/websites/{id}", loginRequired, editWebsiteView)
	app.Get(APP_PATH+"/tracker.png", collectImgView)
	app.Get(APP_PATH+"/{id}", loginRequired, websiteView)

	// POST view handlers
	app.Post(APP_PATH+"/sign-in", signInPostView)
	app.Post(APP_PATH+"/sign-up", signUpPostView)
	app.Post(APP_PATH+"/websites/new", newWebsitePostView)
	app.Post(APP_PATH+"/websites/{id}", loginRequired, editWebsitePostView)
	app.Post(APP_PATH+"/account", loginRequired, changePasswordPost)
	app.Post(APP_PATH+"/{id}", loginRequired, changeDateRangeView)

	sendTestMail()

	app.Listen(conf.ListenAddr)
}
