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
		redirectUrl = conf.AppUrl + APP_PATH + "/websites/new"
	} else {
		redirectUrl = conf.AppUrl + APP_PATH + "/" + strconv.Itoa(int(w.ID))
	}
	ctx.Redirect(redirectUrl)
	return
}

type Website struct {
	gorm.Model
	Owner        *User
	OwnerID      uint   `sql:"index"`
	Name         string `sql:"size:255"`
	Url          string `sql:"size:255"`
	Default      bool   `sql:"not null"`
	TrackingCode string `sql:"size:255"`
}

func (w *Website) getPageViews(older, newer *time.Time) []*PageView {
	var pvs []*PageView
	db.Order("id").Where("website_id = ? AND created_at BETWEEN ? and ?", w.ID, older, newer).Find(&pvs)
	return pvs
}

func (w *Website) getVisitPageViews(older, newer *time.Time) []*PageView {
	var vpvs []*PageView
	gpvs := w.getGroupedPageViews(older, newer)
	for _, pv := range gpvs {
		vpvs = append(vpvs, pv[0])
	}
	return vpvs
}

func (w *Website) getGroupedPageViews(older, newer *time.Time) [][]*PageView {
	var gpvs [][]*PageView
	pvs := w.getPageViews(older, newer)
	push := true
	for i, pv := range pvs {
		first := false
		if i == 0 {
			pvBefore := &PageView{}
			db.Order("id desc").Where("id < ?", pv.ID).First(pvBefore)
			if pvBefore.ID == 0 {
				first = true
			} else {
				d := getDuration(&pvBefore.CreatedAt, &pv.CreatedAt)
				if d.Minutes() >= 30 {
					first = true
				} else {
					push = false
				}
			}
		} else {
			d := getDuration(&pvs[i-1].CreatedAt, &pv.CreatedAt)
			if d.Minutes() >= 30 {
				first = true
			}
		}

		if first {
			newGroup := []*PageView{pv}
			gpvs = append(gpvs, newGroup)
		} else if push {
			gpvs[len(gpvs)-1] = append(gpvs[len(gpvs)-1], pv)
		}
	}
	return gpvs
}

func (w *Website) countPageViews(older, newer *time.Time) int {
	count := 0
	gpvs := w.getGroupedPageViews(older, newer)
	for _, gpv := range gpvs {
		count += len(gpv)
	}
	return count
}

func (w *Website) countVisitors(older, newer *time.Time) int {
	gpvs := w.getGroupedPageViews(older, newer)
	return len(gpvs)
}

func (w *Website) countVisits(older, newer *time.Time) int {
	vpvs := w.getVisitPageViews(older, newer)
	return len(vpvs)
}

func (w *Website) countBouncedVisits(older, newer *time.Time) int {
	count := 0
	gpvs := w.getGroupedPageViews(older, newer)
	for _, gpv := range gpvs {
		if len(gpv) == 1 {
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

func (w *Website) getDataPoints(numDays, limit int) []int {
	var dataPoints []int
	for ; limit > 0; limit-- {
		dataPoints = append(dataPoints, w.countVisits(getTimeDaysAgo(numDays), getTimeDaysAgo(numDays-1)))
		numDays--
	}
	return dataPoints
}

func (w *Website) getDataPointsHourly(numDays int) []int {
	var dataPoints []int
	start := getTimeDaysAgo(numDays + 1)
	for i := 0; i < 24; i++ {
		older := start.Add(time.Duration(i) * time.Hour)
		newer := start.Add(time.Duration(i+1) * time.Hour).Add(-time.Second)
		dataPoints = append(dataPoints, w.countVisits(&older, &newer))
	}
	return dataPoints
}

func (w *Website) getTimePerVisit(older, newer *time.Time) string {
	seconds := 0

	gpvs := w.getGroupedPageViews(older, newer)
	for _, gpv := range gpvs {
		if len(gpv) > 1 {
			sinceOlder := time.Since(gpv[0].CreatedAt)
			sinceNewer := time.Since(gpv[len(gpv)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	var d time.Duration

	if len(gpvs) > 0 {
		d = time.Duration(time.Second * time.Duration(seconds/len(gpvs)))
	} else {
		d = time.Duration(0)
	}

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (w *Website) getTimeAllVisits(older, newer *time.Time) string {
	seconds := 0

	gpvs := w.getGroupedPageViews(older, newer)
	for _, gpv := range gpvs {
		if len(gpv) > 1 {
			sinceOlder := time.Since(gpv[0].CreatedAt)
			sinceNewer := time.Since(gpv[len(gpv)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	d := time.Duration(time.Second * time.Duration(seconds))

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (w *Website) getPageViewsPerVisit(older, newer *time.Time) string {
	count := 0
	gpvs := w.getGroupedPageViews(older, newer)
	for _, gpv := range gpvs {
		count += len(gpv)
	}
	return fmt.Sprintf("%.2f", float64(count)/float64(len(gpvs)))
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
	Visit   *Visit
	VisitID uint `sql:"index"`
	Page    *Page
	PageID  uint `sql:"index"`
}
