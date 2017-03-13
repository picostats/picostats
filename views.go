package main

import (
	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": iris.NoLayout})
}

func signInPostView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": iris.NoLayout})
}

func signOutView(ctx *iris.Context) {
	ctx.Writef("Hi %s", "iris")
}

func signUpView(ctx *iris.Context) {
	ctx.Writef("Hi %s", "iris")
}

func dashboardView(ctx *iris.Context) {
	ctx.Session().Set("aaa", true)

	ctx.Writef("Hi %s", "data")
}
