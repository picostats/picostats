package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"

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

func getTrackerImageBytes() []byte {
	infile, err := os.Open(TRACKER_IMAGE)
	if err != nil {
		log.Printf("[views.go] Image open error: %s", err)
	}
	defer infile.Close()

	fileInfo, _ := infile.Stat()
	var size int64 = fileInfo.Size()
	bytes := make([]byte, size)

	buffer := bufio.NewReader(infile)
	_, err = buffer.Read(bytes)
	return bytes
}

type PageViewRequest struct {
	WebsiteID  string `json:"website_id"`
	Title      string `json:"title,omitempty"`
	Path       string `json:"path,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Language   string `json:"language,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Referrer   string `json:"referrer,omitempty"`
}
