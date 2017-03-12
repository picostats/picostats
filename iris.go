package main

import (
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
)

func initIris() {
	app = iris.New()
	app.Adapt(httprouter.New())
}
