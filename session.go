package main

import (
	"time"

	"gopkg.in/kataras/iris.v6/adaptors/sessions"
	"gopkg.in/kataras/iris.v6/adaptors/sessions/sessiondb/redis"
	"gopkg.in/kataras/iris.v6/adaptors/sessions/sessiondb/redis/service"
)

func initSession() {
	session := sessions.New(sessions.Config{
		Cookie:                      SESSION_COOKIE,
		DecodeCookie:                false,
		Expires:                     time.Hour * 2,
		CookieLength:                32,
		DisableSubdomainPersistence: false,
	})

	rConf := service.Config{Addr: conf.RedisUrl}

	session.UseDatabase(redis.New(rConf))

	app.Adapt(session)
}
