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
	session.Set(USER_ID, user.ID)
}

func signOut(ctx *iris.Context) {
	session := ctx.Session()
	session.Delete(USER_ID)
}

func isSignedIn(ctx *iris.Context) bool {
	session := ctx.Session()
	userId := session.Get(USER_ID)
	if userId == nil {
		return false
	}
	if userId.(uint) > 0 {
		u := &User{}
		db.First(u, userId.(uint))
		return u.ID > 0 && u.Verified
	}
	return false
}
