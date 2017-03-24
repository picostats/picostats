package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

type Worker struct{}

func (w *Worker) work() {
	for {
		hasKay, _ := red.Exists("pvs").Result()
		if hasKay {
			pvJson := red.LPop("pvs").Val()
			pvr := &PageViewRequest{}
			err := json.Unmarshal([]byte(pvJson), pvr)
			if err != nil {
				log.Printf("[worker.go] Error in unarshall: %s", err)
			}
			w.handlePageViewRequest(pvr)
		} else {
			time.Sleep(time.Millisecond * 200)
		}
	}
}

func (w *Worker) workReports() {
	reportTypes := []int{1, 2, 3, 4, 5, 6}
	for {
		var websites []*Website
		db.Find(&websites)
		for _, w := range websites {
			u := &User{}
			db.First(u, w.OwnerID)
			for _, rt := range reportTypes {
				start, end := rm.getDefaultTimes(u.TimeOffset, rt)
				repMod := &ReportModel{WebsiteID: w.ID, Type: rt}
				db.First(repMod, repMod)

				if repMod.ID == 0 || time.Since(repMod.UpdatedAt).Seconds() > 60 {
					newRep := rm.generateNew(rt, int(start.Unix()), int(end.Unix()), w, u.TimeOffset)

					repMod.Visits = newRep.Visits
					repMod.Visitors = newRep.Visitors
					repMod.PageViews = newRep.PageViews
					repMod.BounceRate = newRep.BounceRate
					repMod.New = newRep.New
					repMod.Returning = newRep.Returning
					repMod.DataPoints = joinDataPoints(newRep.DataPoints)
					repMod.DataPointsPast = joinDataPoints(newRep.DataPointsPast)
					repMod.TimePerVisit = newRep.TimePerVisit
					repMod.TimeTotal = newRep.TimeTotal
					repMod.PageViewsPerVisit = newRep.PageViewsPerVisit
					repMod.NewPercentage = newRep.NewPercentage
					repMod.ReturningPercentage = newRep.ReturningPercentage
					repMod.DateRangeType = newRep.DateRangeType
					repMod.ChartScale = strings.Join(newRep.ChartScale, "|")
					repMod.StartInt = start.Unix()
					repMod.EndInt = end.Unix()

					db.Save(repMod)
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func (w *Worker) handlePageViewRequest(pvr *PageViewRequest) {
	var vNew *Visit
	website := &Website{}
	db.Where("tracking_code = ?", pvr.WebsiteID).First(website)

	visitor := &Visitor{
		IpAddress:  pvr.IpAddress,
		Resolution: pvr.Resolution,
		Language:   pvr.Language,
	}
	db.FirstOrCreate(visitor, visitor)

	page := &Page{
		Hostname: pvr.Hostname,
		Path:     pvr.Path,
		Title:    pvr.Title,
	}
	db.FirstOrCreate(page, page)

	v := &Visit{
		WebsiteID: website.ID,
		VisitorID: visitor.ID,
	}
	db.Order("id desc").Where(v).First(v)

	if v.ID != 0 {
		pv := &PageView{
			VisitID: v.ID,
		}
		db.Order("id desc").Where(pv).First(pv)

		delta := time.Now().Sub(pv.CreatedAt)
		if delta.Minutes() > 30 {
			vNew = &Visit{
				WebsiteID: website.ID,
				VisitorID: visitor.ID,
			}
			db.Create(vNew)
		} else {
			vNew = v
		}
	} else {
		vNew = &Visit{
			WebsiteID: website.ID,
			VisitorID: visitor.ID,
		}
		db.Create(vNew)
	}

	pvNew := &PageView{
		VisitID:        vNew.ID,
		PageID:         page.ID,
		WebsiteID:      website.ID,
		SignedInUserId: pvr.SignedInUserId,
	}
	db.Create(pvNew)
}

func initWorker() {
	w := Worker{}
	if clip.Command == "worker" {
		go w.work()
		w.workReports()
	} else if clip.Command == "server" {
		go w.work()
		go w.workReports()
	}
}
