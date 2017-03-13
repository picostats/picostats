package main

import (
	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	name, _ := ctx.Session().GetBoolean("aaa")

	ctx.Writef("Hi %s", name)
}

func signUpView(ctx *iris.Context) {
	ctx.Writef("Hi %s", "iris")
}

func dashboardView(ctx *iris.Context) {
	ctx.Session().Set("aaa", true)

	ctx.Writef("Hi %s", "data")
}
