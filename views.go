package main

import (
	"log"

	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": iris.NoLayout})
}

func signInPostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	sif := &SignInForm{}
	err := ctx.ReadForm(sif)
	if err != nil {
		log.Println("[views.go] Error reading SignInForm: %s", err)
	}

	user := &User{Email: sif.Email}
	db.First(user)

	log.Println(user.ID)

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
