package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

type Worker struct{}

func (w *Worker) work() {
	go func() {
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
	}()
}

func (w *Worker) handlePageViewRequest(pvr *PageViewRequest) {
	website := &Website{}
	wId, _ := strconv.Atoi(aesDecrypt(pvr.WebsiteID))
	db.First(website, wId)

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

	pv := &PageView{
		WebsiteID: website.ID,
		VisitorID: visitor.ID,
		PageID:    page.ID,
	}

	db.Create(pv)
}

func initWorker() {
	w := Worker{}
	w.work()
}
