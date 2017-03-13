package main

import (
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
)

func initIris() *iris.Framework {
	app := iris.New()
	app.Adapt(httprouter.New())
	app.Adapt(view.HTML("./templates", ".html"))
	app.StaticWeb("/public", "./public")
	return app
}
