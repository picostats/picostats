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
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/kataras/iris.v6"
)

type PageData struct {
	User                   *User
	Websites               []*Website
	Conf                   *Config
	Errors                 []*error
	Form                   interface{}
	Gravatar               string
	WebsiteId              string
	TrackerUrl             string
	SuccessFlash           interface{}
	ErrorFlash             interface{}
	Report                 *Report
	DataRangeStartSubtract int
	DataRangeEndSubract    int
	DateRangeType          string
	ChartScale             []string
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

type Report struct {
	Visits              int
	Visitors            int
	PageViews           int
	BounceRate          string
	New                 int
	Returning           int
	DataPoints          []int
	DataPointsPast      []int
	TimePerVisit        string
	TimeTotal           string
	PageViewsPerVisit   string
	NewPercentage       string
	ReturningPercentage string
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
		pd.Websites = websites
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

func getTimeDaysAgo(numDays int) *time.Time {
	numDays--
	timeAgo := time.Now().Truncate(time.Hour).Add(-time.Hour*time.Duration(time.Now().Hour())).AddDate(0, 0, -numDays)
	return &timeAgo
}

func getDateRangeType(startSubtract, endSubract int) string {
	dateRangeType := "Date Range"
	if startSubtract == 0 && endSubract == 0 {
		dateRangeType = "Today"
	} else if startSubtract == 1 && endSubract == 1 {
		dateRangeType = "Yesterday"
	} else if startSubtract == 6 && endSubract == 0 {
		dateRangeType = "Last 7 Days"
	} else if startSubtract == 29 && endSubract == 0 {
		dateRangeType = "Last 30 Days"
	} else if endSubract == 0 {
		dateRangeType = "This Month"
	} else {
		dateRangeType = "Last Month"
	}
	return dateRangeType
}

func getChartScale(startSubtract, endSubract int) []string {
	chartScale := []string{}
	if (startSubtract == 0 && endSubract == 0) || (startSubtract == 1 && endSubract == 1) {
		chartScale = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	} else if startSubtract == 6 && endSubract == 0 {
		chartScale = []string{}
		for i := -6; i <= 0; i++ {
			item := time.Now().AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(time.Now().AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	} else if startSubtract == 29 && endSubract == 0 {
		chartScale = []string{}
		for i := -29; i <= 0; i++ {
			item := time.Now().AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(time.Now().AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	} else if endSubract == 0 {
		timeCounter := time.Now().AddDate(0, 0, -time.Now().Day()+1)
		for timeCounter.Month() == time.Now().Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	} else {
		timeCounterMonth := time.Now().AddDate(0, -1, 0)
		timeCounter := timeCounterMonth.AddDate(0, 0, -timeCounterMonth.Day()+1)
		for timeCounter.Month() == timeCounterMonth.Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	}
	return chartScale
}

func appPath() string {
	u, err := url.Parse(conf.AppUrl)
	if err != nil {
		log.Printf("[helpers.go] Error parsing URL: %s", err)
	}
	return u.Path
}
