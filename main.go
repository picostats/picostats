package main

import (
    "log"

    "gopkg.in/kataras/iris.v6"
    "gopkg.in/kataras/iris.v6/adaptors/httprouter"
)

var conf *config

func main() {
    app := iris.New()
    app.Adapt(httprouter.New())

    initConfig()

    log.Println(conf.RedisUrl)

    app.Get("/", home)

    app.Listen(":8080")
}