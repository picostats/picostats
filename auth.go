package main

import (
	"gopkg.in/kataras/iris.v6"
)

func loginRequired(ctx *iris.Context) {
	ctx.Next()
}

func signIn(ctx *iris.Context) {

}

func signOut(ctx *iris.Context) {

}
