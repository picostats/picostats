package main

import (
    "gopkg.in/kataras/iris.v6"
    "gopkg.in/kataras/iris.v6/adaptors/httprouter"
)

func main() {
    app := iris.New()
    app.Adapt(httprouter.New())

    app.Get("/", func(ctx *iris.Context) {
        ctx.Writef("Hi %s", "iris")
    })

    app.Listen(":8080")
}