package main

import (
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
)

func initIris() *iris.Framework {
	app := iris.New()
	app.Adapt(httprouter.New())
	return app
}
