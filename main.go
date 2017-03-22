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

var em *EmailManager

var rm *ReportManager

var tzm *TimeZonesManager

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

	// Initializes time zones parser
	tzm = initZones()

	// Initializes email service
	initEmails()

	// Initializes worker and starts saving data
	initWorker()

	// Initializes report manager for generating reports
	initReport()

	// GET view handlers
	app.Get(appPath(), redirectView)
	app.Get(appPath()+"/sign-in", signInView)
	app.Get(appPath()+"/install", installView)
	app.Get(appPath()+"/install2", installView2)
	app.Get(appPath()+"/sign-up", signUpView)
	app.Get(appPath()+"/sign-out", signOutView)
	app.Get(appPath()+"/account", loginRequired, accountView)
	app.Get(appPath()+"/account/delete", loginRequired, accountDeleteView)
	app.Get(appPath()+"/websites/new", loginRequired, newWebsiteView)
	app.Get(appPath()+"/websites/delete/{id}", loginRequired, websiteDeleteView)
	app.Get(appPath()+"/websites/default/{id}", loginRequired, websiteMakeDefaultView)
	app.Get(appPath()+"/websites/{id}", loginRequired, editWebsiteView)
	app.Get(appPath()+"/verify", verifyMessageView)
	app.Get(appPath()+"/verify/{hash}", verifyView)
	app.Get(appPath()+"/tracker.png", collectImgView)
	app.Get(appPath()+"/{id}", loginRequired, websiteView)

	// POST view handlers
	app.Post(appPath()+"/sign-in", signInPostView)
	app.Post(appPath()+"/install", installPostView)
	app.Post(appPath()+"/sign-up", signUpPostView)
	app.Post(appPath()+"/websites/new", loginRequired, newWebsitePostView)
	app.Post(appPath()+"/websites/{id}", loginRequired, editWebsitePostView)
	app.Post(appPath()+"/account", loginRequired, changePasswordPost)
	app.Post(appPath()+"/account/settings", loginRequired, saveSettingsPostView)
	app.Post(appPath()+"/{id}", loginRequired, changeDateRangeView)

	app.Listen(conf.ListenAddr)
}
