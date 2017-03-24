package main

import (
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/gorillamux"
	"gopkg.in/kataras/iris.v6/adaptors/view"
)

func initIris() *iris.Framework {
	app := iris.New()
	app.Adapt(gorillamux.New())
	app.StaticWeb(appPath()+"/public", "./public")
	if conf.Dev {
		app.Adapt(iris.DevLogger())
		app.Adapt(view.HTML("./templates", ".html").Layout("layout.html").Reload(true))
		app.Adapt(view.HTML("./templates", ".js").Reload(true))
	} else {
		app.Adapt(view.HTML("./templates", ".html").Layout("layout.html"))
		app.Adapt(view.HTML("./templates", ".js"))
	}
	return app
}
