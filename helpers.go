package main

import (
	"crypto/md5"
	"encoding/hex"

	"gopkg.in/kataras/iris.v6"
)

type PageData struct {
	User   *User
	Conf   *Config
	Errors []*error
	Form   interface{}
}

func newPageData(ctx *iris.Context) *PageData {
	pd := &PageData{}
	pd.Conf = conf
	if isSignedIn(ctx) {
		session := ctx.Session()
		userId := session.Get(USER_ID)
		pd.User = &User{}
		db.First(pd.User, userId.(uint))
	}
	return pd
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
