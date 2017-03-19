package main

import (
	"fmt"
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
		redirectUrl = conf.AppUrl + "/websites/new"
	} else {
		redirectUrl = conf.AppUrl + "/" + strconv.Itoa(int(w.ID))
	}
	ctx.Redirect(redirectUrl)
	return
}

func (u *User) countWebsites() int {
	var websites []*Website
	var cnt int
	db.Where("owner_id = ?", u.ID).Find(&websites).Count(&cnt)
	return cnt
}

type Website struct {
	gorm.Model
	Owner        *User
	OwnerID      uint   `sql:"index"`
	Name         string `sql:"size:255"`
	Url          string `sql:"size:255"`
	Default      bool   `sql:"not null"`
	TrackingCode string `sql:"size:255;unique_index"`
}

func (w *Website) countPageViews(older, newer *time.Time) int {
	pvs := w.getPageViews(older, newer)
	return len(pvs)
}

func (w *Website) getPageViews(older, newer *time.Time) []*PageView {
	var pvs []*PageView
	db.Order("id").Where("website_id = ? AND created_at BETWEEN ? AND ?", w.ID, older, newer).Find(&pvs)
	return pvs
}

func (w *Website) countVisitors(older, newer *time.Time) int {
	visitors := map[uint]bool{}
	visits := w.getVisits(older, newer)
	for _, v := range visits {
		visitors[v.VisitorID] = true
	}
	return len(visitors)
}

func (w *Website) countBouncedVisits(older, newer *time.Time) int {
	count := 0

	visits := w.getVisits(older, newer)
	for _, v := range visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		if cnt == 1 {
			count++
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
	return w.countVisitors(older, newer)
}

func (w *Website) countReturning(older, newer *time.Time) int {
	newCount := w.countNew(older, newer)
	visits := w.countVisits(older, newer)
	return visits - newCount
}

func (w *Website) getDataPoints(numDays, limit int, ctx *iris.Context) []int {
	var dataPoints []int
	limitStart := limit
	for ; limit > 0; limit-- {
		if limitStart == limit {
			dataPoints = append(dataPoints, w.countVisits(getTimeDaysAgo(numDays, ctx), getTimeDaysAgo(numDays-1, ctx)))
		} else {
			dataPoints = append(dataPoints, w.countVisitsPrecise(getTimeDaysAgo(numDays, ctx), getTimeDaysAgo(numDays-1, ctx)))
		}
		numDays--
	}
	return dataPoints
}

func (w *Website) getDataPointsHourly(numDays int, ctx *iris.Context) []int {
	var dataPoints []int
	start := getTimeDaysAgo(numDays+1, ctx)
	for i := 0; i < 24; i++ {
		older := start.Add(time.Duration(i) * time.Hour)
		newer := start.Add(time.Duration(i+1) * time.Hour).Add(-time.Second)
		if i == 0 {
			dataPoints = append(dataPoints, w.countVisits(&older, &newer))
		} else {
			dataPoints = append(dataPoints, w.countVisitsPrecise(&older, &newer))
		}
	}
	return dataPoints
}

func (w *Website) getVisitsPrecise(older, newer *time.Time) []*Visit {
	var visits []*Visit
	db.Order("id").Where("website_id = ? AND created_at BETWEEN ? AND ?", w.ID, older, newer).Find(&visits)
	return visits
}

func (w *Website) getVisits(older, newer *time.Time) []*Visit {
	var visits []*Visit
	visitsTemp := make(map[uint]*Visit)
	pvs := w.getPageViews(older, newer)
	for _, pv := range pvs {
		_, ok := visitsTemp[pv.VisitID]
		if !ok {
			v := &Visit{}
			db.First(v, pv.VisitID)
			visits = append(visits, v)
			visitsTemp[pv.VisitID] = v
		}
	}
	return visits
}

func (w *Website) countVisits(older, newer *time.Time) int {
	visits := w.getVisits(older, newer)
	return len(visits)
}

func (w *Website) countVisitsPrecise(older, newer *time.Time) int {
	visits := w.getVisitsPrecise(older, newer)
	return len(visits)
}

func (w *Website) getTimePerVisit(older, newer *time.Time) string {
	seconds := 0

	visits := w.getVisits(older, newer)
	for _, v := range visits {
		var pvs []*PageView
		db.Order("id").Where(&PageView{VisitID: v.ID}).Find(&pvs)
		if len(pvs) > 1 {
			sinceOlder := time.Since(pvs[0].CreatedAt)
			sinceNewer := time.Since(pvs[len(pvs)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	var d time.Duration

	if len(visits) > 0 {
		d = time.Duration(time.Second * time.Duration(seconds/len(visits)))
	} else {
		d = time.Duration(0)
	}

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (w *Website) getTimeAllVisits(older, newer *time.Time) string {
	seconds := 0

	visits := w.getVisits(older, newer)
	for _, v := range visits {
		var pvs []*PageView
		db.Order("id").Where(&PageView{VisitID: v.ID}).Find(&pvs)
		if len(pvs) > 1 {
			sinceOlder := time.Since(pvs[0].CreatedAt)
			sinceNewer := time.Since(pvs[len(pvs)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	d := time.Duration(time.Second * time.Duration(seconds))

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (w *Website) getPageViewsPerVisit(older, newer *time.Time) string {
	count := 0

	visits := w.getVisits(older, newer)
	for _, v := range visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		count += cnt
	}

	return fmt.Sprintf("%.2f", float64(count)/float64(len(visits)))
}

type Page struct {
	gorm.Model
	Hostname string `sql:"size:255"`
	Path     string `sql:"size:255"`
	Title    string `sql:"size:255"`
}

type Visitor struct {
	gorm.Model
	IpAddress  string `sql:"size:255"`
	Resolution string `sql:"size:255"`
	Language   string `sql:"size:255"`
}

type Visit struct {
	gorm.Model
	Website   *Website
	WebsiteID uint `sql:"index"`
	Visitor   *Visitor
	VisitorID uint `sql:"index"`
}

type PageView struct {
	gorm.Model
	Visit     *Visit
	VisitID   uint `sql:"index"`
	Page      *Page
	PageID    uint `sql:"index"`
	Website   *Website
	WebsiteID uint `sql:"index";default=1`
}
