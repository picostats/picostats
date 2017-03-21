package main

import (
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/kataras/iris.v6"
)

type User struct {
	gorm.Model
	Email       string `sql:"size:255" unique_index`
	Password    string `sql:"size:255"`
	Verified    bool   `sql:"not null"`
	ExcludeMe   bool   `sql:"not null"`
	MaxWebsites int    `sql`
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

func (w *Website) getPageViews(older, newer *time.Time) []*PageView {
	var pvs []*PageView
	u := &User{}
	db.First(u, w.OwnerID)
	if u.ExcludeMe {
		db.Order("id").Where("signed_in_user_id IS DISTINCT FROM ? AND website_id = ? AND created_at BETWEEN ? AND ?", u.ID, w.ID, older, newer).Find(&pvs)
	} else {
		db.Order("id").Where("website_id = ? AND created_at BETWEEN ? AND ?", w.ID, older, newer).Find(&pvs)
	}
	return pvs
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
	Visit          *Visit
	VisitID        uint `sql:"index"`
	Page           *Page
	PageID         uint `sql:"index"`
	Website        *Website
	WebsiteID      uint `sql:"index"`
	SignedInUserId uint `sql:"index"`
}

type ModelReport struct {
	Website   *Website
	WebsiteID uint `sql:"index"`

	Type uint `sql`

	Visits              uint   `sql`
	Visitors            uint   `sql`
	PageViews           uint   `sql`
	BounceRate          string `sql:"size:255"`
	New                 uint   `sql`
	Returning           uint   `sql`
	DataPoints          string `sql:"size:255"`
	DataPointsPast      string `sql:"size:255"`
	TimePerVisit        string `sql:"size:255"`
	TimeTotal           string `sql:"size:255"`
	PageViewsPerVisit   string `sql:"size:255"`
	NewPercentage       string `sql:"size:255"`
	ReturningPercentage string `sql:"size:255"`
	DateRangeType       string `sql:"size:255"`
	ChartScale          string `sql:"size:255"`
	StartInt            uint   `sql`
	EndInt              uint   `sql`
}
