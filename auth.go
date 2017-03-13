package main

import (
	"gopkg.in/kataras/iris.v6"
)

func loginRequired(ctx *iris.Context) {
	if !isSignedIn(ctx) {
		ctx.Redirect(conf.AppUrl + APP_PATH + "/sign-in")
	} else {
		ctx.Next()
	}
}

func signIn(ctx *iris.Context, user *User) {
	session := ctx.Session()
	session.Set("userid", user.ID)
}

func signOut(ctx *iris.Context) {
	session := ctx.Session()
	session.Delete("userid")
}

func isSignedIn(ctx *iris.Context) bool {
	session := ctx.Session()
	userId := session.Get("userid")
	if userId == nil {
		return false
	}
	if userId.(uint) > 0 {
		return true
	}
	return false
}
