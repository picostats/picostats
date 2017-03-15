package main

import (
	// "log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
)

type User struct {
	gorm.Model
	Email    string `sql:"size:255" unique_index`
	Password string `sql:"size:255"`
	Verified bool   `sql:"not null"`
}

func (u *User) getDefaultWebsite() *Website {
	w := &Website{OwnerID: u.ID, Default: true}
	db.Where(w).First(w)
	return w
}

func (u *User) redirectToDefaultWebsite(ctx *iris.Context) {
	w := u.getDefaultWebsite()
	var redirectUrl string
	if w.ID == 0 {
		redirectUrl = conf.AppUrl + APP_PATH + "/websites/new"
	} else {
		redirectUrl = conf.AppUrl + APP_PATH + "/" + strconv.Itoa(int(w.ID))
	}
	ctx.Redirect(redirectUrl)
	return
}

type Website struct {
	gorm.Model
	Owner   *User
	OwnerID uint   `sql:"index"`
	Name    string `sql:"size:255"`
	Url     string `sql:"size:255"`
	Default bool   `sql:"not null"`
}

func (w *Website) getPageViews(older, newer *time.Time) []*PageView {
	var pvs []*PageView
	db.Order("id").Where("website_id = ? AND created_at BETWEEN ? and ?", w.ID, older, newer).Find(&pvs)
	return pvs
}

func (w *Website) countPageViews(older, newer *time.Time) int {
	pvs := w.getPageViews(older, newer)
	return len(pvs)
}

func (w *Website) countUsers(older, newer *time.Time) int {
	counter := map[uint]bool{}
	pvs := w.getPageViews(older, newer)
	for _, pv := range pvs {
		counter[pv.VisitorID] = true
	}
	return len(counter)
}

func (w *Website) countVisits(older, newer *time.Time) int {
	count := 0
	pvs := w.getPageViews(older, newer)
	for i, pv := range pvs {
		if i < len(pvs)-1 {
			d := getDuration(&pv.CreatedAt, &pvs[i+1].CreatedAt)
			if d.Minutes() >= 30 {
				count++
			}
		}
		if i == len(pvs)-1 {
			count++
		}
	}
	return count
}

func (w *Website) countBouncedVisits(older, newer *time.Time) int {
	count := 0
	pvs := w.getPageViews(older, newer)
	for i, pv := range pvs {
		if i < len(pvs)-1 {
			d := getDuration(&pv.CreatedAt, &pvs[i+1].CreatedAt)
			if d.Minutes() >= 30 {
				if i == 0 {
					count++
				} else {
					d := getDuration(&pv.CreatedAt, &pvs[i-1].CreatedAt)
					if d.Minutes() >= 30 {
						count++
					}
				}
			}
		}

	}
	return count
}

func (w *Website) getBounceRate(older, newer *time.Time) float64 {
	visits := w.countVisits(older, newer)
	if visits > 0 {
		blounceRate := float64(w.countBouncedVisits(older, newer)) / float64(visits) * float64(100)
		return blounceRate
	}
	return 0
}

func (w *Website) countNew(older, newer *time.Time) int {
	return w.countUsers(older, newer)
}

func (w *Website) countReturning(older, newer *time.Time) int {
	newCount := w.countNew(older, newer)
	visits := w.countVisits(older, newer)
	return visits - newCount
}

func (w *Website) getDataPoints(numDays, limit int) []int {
	var dataPoints []int
	for ; limit > 0; limit-- {
		dataPoints = append(dataPoints, w.countVisits(getTimeDaysAgo(numDays), getTimeDaysAgo(numDays-1)))
		numDays--
	}
	// log.Println(dataPoints)
	return dataPoints
}

type Visitor struct {
	gorm.Model
	IpAddress  string `sql:"size:255"`
	Resolution string `sql:"size:255"`
	Language   string `sql:"size:255"`
}

type Page struct {
	gorm.Model
	Hostname string `sql:"size:255"`
	Path     string `sql:"size:255"`
	Title    string `sql:"size:255"`
}

type PageView struct {
	gorm.Model
	Website   *Website
	WebsiteID uint `sql:"index"`
	Visitor   *Visitor
	VisitorID uint `sql:"index"`
	Page      *Page
	PageID    uint `sql:"index"`
}
