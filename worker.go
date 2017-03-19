package main

import (
	"encoding/json"
	"log"
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
		VisitID:   vNew.ID,
		PageID:    page.ID,
		WebsiteID: website.ID,
	}
	db.Create(pvNew)
}

func initWorker() {
	w := Worker{}
	if clip.Command == "worker" {
		w.work()
	} else if clip.Command == "server" {
		go w.work()
	}
}
