package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/kataras/iris.v6"
)

type PageData struct {
	User     *User
	Conf     *Config
	Errors   []*error
	Form     interface{}
	Gravatar string
}

func newPageData(ctx *iris.Context) *PageData {
	pd := &PageData{}
	pd.Conf = conf
	if isSignedIn(ctx) {
		session := ctx.Session()
		userId := session.Get(USER_ID)
		pd.User = &User{}
		db.First(pd.User, userId.(uint))
		if conf.Dev {
			pd.Gravatar = conf.AppUrl + "/public/img/user.png"
		} else {
			placeholder := conf.AppUrl + "/public/img/user.png"
			placeholder = strings.Replace(placeholder, ":", "%3A", -1)
			placeholder = strings.Replace(placeholder, "/", "%2F", -1)
			pd.Gravatar = fmt.Sprintf("https://secure.gravatar.com/avatar/%x?s=50&d=%s", md5.Sum([]byte(pd.User.Email)), placeholder)
		}
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
	IpAddress  string `json:"ip_address,omitempty"`
}

func aesEncrypt(text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(conf.EncryptionKey)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext)
}

func aesDecrypt(cryptoText string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(conf.EncryptionKey)
	if err != nil {
		panic(err)
	}

	if len(ciphertext) < aes.BlockSize {
		log.Printf("[AesDecrypt] ciphertext too short")
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}
