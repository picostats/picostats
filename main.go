package main

import (
    "gopkg.in/kataras/iris.v6"
    "gopkg.in/kataras/iris.v6/adaptors/httprouter"
)

func main() {
    app := iris.New()
    app.Adapt(httprouter.New())

    app.Get("/", home)

    app.Listen(":8080")
}