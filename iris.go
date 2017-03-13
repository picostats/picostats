package main

import (
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
)

func initIris() *iris.Framework {
	app := iris.New()
	app.Adapt(httprouter.New())
	app.StaticWeb("/public", "./public")
	if conf.Dev {
		app.Adapt(iris.DevLogger())
		app.Adapt(view.HTML("./templates", ".html").Layout("layout.html").Reload(true))
	} else {
		app.Adapt(view.HTML("./templates", ".html").Layout("layout.html"))
	}
	return app
}
