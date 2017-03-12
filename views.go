package main

import (
    "gopkg.in/kataras/iris.v6"
)

func home(ctx *iris.Context) {
    ctx.Writef("Hi %s", "iris")
}
