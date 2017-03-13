package main

import (
	"crypto/md5"
	"encoding/hex"

	"gopkg.in/kataras/iris.v6"
)

type PageData struct {
	User *User
	Conf *Config
}

func newPageData(ctx *iris.Context) *PageData {
	pd := &PageData{}
	return pd
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
