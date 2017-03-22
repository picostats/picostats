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
	"html/template"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/kataras/iris.v6"
)

type PageData struct {
	User         *User
	Websites     []*Website
	Conf         *Config
	Errors       []*error
	Form         interface{}
	Gravatar     string
	WebsiteId    string
	TrackerUrl   string
	SuccessFlash interface{}
	ErrorFlash   interface{}
	Report       *Report
	TitlePrefix  string
	TimeZones    []string
}

type PageViewRequest struct {
	WebsiteID      string `json:"website_id"`
	Title          string `json:"title,omitempty"`
	Path           string `json:"path,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Language       string `json:"language,omitempty"`
	Resolution     string `json:"resolution,omitempty"`
	Referrer       string `json:"referrer,omitempty"`
	IpAddress      string `json:"ip_address,omitempty"`
	SignedInUserId uint   `json:"signed_in_user_id,omitempty"`
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
		var websites []*Website
		db.Order("id").Where("owner_id = ?", pd.User.ID).Find(&websites)
		if pd.User.countWebsites() >= pd.User.MaxWebsites && pd.User.MaxWebsites != 0 {
			for _, w := range websites {
				if w.Default {
					pd.Websites = []*Website{w}
				}
			}
		} else {
			pd.Websites = websites
		}
	}
	session := ctx.Session()
	sFl := session.GetFlash("success")
	eFl := session.GetFlash("error")
	if sFl != nil {
		pd.SuccessFlash = template.HTML(sFl.(string))
	}
	if eFl != nil {
		pd.ErrorFlash = template.HTML(eFl.(string))
	}
	if conf.AppUrl == "/" {
		conf.AppUrl = ""
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

func getDuration(older, newer *time.Time) *time.Duration {
	sinceOlder := time.Since(*older)
	sinceNewer := time.Since(*newer)
	minutes := sinceOlder.Minutes() - sinceNewer.Minutes()
	d := time.Duration(time.Minute * time.Duration(minutes))
	return &d
}

func appPath() string {
	u, err := url.Parse(conf.AppUrl)
	if err != nil {
		log.Printf("[helpers.go] Error parsing URL: %s", err)
	}
	if u.Path == "/" {
		u.Path = ""
	}
	return u.Path
}

func round(val float64) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(0))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= 0.0 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func joinDataPoints(dataPoints []int) string {
	dataPointsStr := ""
	for i, dp := range dataPoints {
		if i == 0 {
			dataPointsStr += strconv.Itoa(dp)
		} else {
			dataPointsStr += "|" + strconv.Itoa(dp)
		}
	}
	return dataPointsStr
}

func splitDataPoints(dataPointsStr string) []int {
	var dataPoints []int
	dataPointsSlice := strings.Split(dataPointsStr, "|")
	for _, dp := range dataPointsSlice {
		dpInt, err := strconv.Atoi(dp)
		if err != nil {
			log.Printf("[helpers.go] Error in Atoi: %s", err)
		}
		dataPoints = append(dataPoints, dpInt)
	}
	return dataPoints
}
