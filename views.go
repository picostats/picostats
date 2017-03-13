package main

import (
	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("login.html", pd)
}

func signInPostView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("login.html", pd)
}

func signUpView(ctx *iris.Context) {
	ctx.Writef("Hi %s", "iris")
}

func dashboardView(ctx *iris.Context) {
	ctx.Session().Set("aaa", true)

	ctx.Writef("Hi %s", "data")
}
