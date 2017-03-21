package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"gopkg.in/kataras/iris.v6"
)

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
	DateRangeType       string
	ChartScale          []string
	StartInt            int64
	EndInt              int64
}

type ReportManager struct {
	PageViews     []*PageView
	Visits        []*Visit
	VisitsPrecise []*Visit
	Website       *Website
}

func (r *ReportManager) generateReport(ctx *iris.Context, w *Website) *Report {
	r.Website = w
	report := &Report{}
	session := ctx.Session()

	startStr := session.GetString("date-range-start")
	endStr := session.GetString("date-range-end")

	if len(startStr) == 0 {
		t := getTimeDaysAgo(7, ctx)
		startStr = strconv.Itoa(int(t.Unix()))
	}
	if len(endStr) == 0 {
		t := getTimeDaysAgo(1, ctx)
		endStr = strconv.Itoa(int(t.Unix()))
	}

	startInt, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		log.Printf("[views.go] Error parsing timestamp: %s", err)
	}
	report.StartInt = startInt
	start := time.Unix(startInt, 0)

	endInt, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		log.Printf("[views.go] Error parsing timestamp: %s", err)
	}
	report.EndInt = endInt
	end := time.Unix(endInt, 0)

	r.PageViews = w.getPageViews(&start, &end)
	r.Visits = w.getVisits(&start, &end)
	r.VisitsPrecise = w.getVisitsPrecise(&start, &end)

	report.DateRangeType = r.getDateRangeType(start, end, ctx)
	report.ChartScale = r.getChartScale(start, end, ctx)

	report.PageViews = len(r.PageViews)
	report.Visitors = r.countVisitors()
	report.Visits = len(r.Visits)
	report.New = r.countNew()
	report.Returning = r.countReturning()
	report.BounceRate = fmt.Sprintf("%.2f", r.getBounceRate())
	report.TimePerVisit = r.getTimePerVisit()
	report.TimeTotal = r.getTimeAllVisits()
	report.PageViewsPerVisit = r.getPageViewsPerVisit()
	report.NewPercentage = fmt.Sprintf("%.2f", float64(report.New)/float64(report.New+report.Returning)*100)
	report.ReturningPercentage = fmt.Sprintf("%.2f", float64(report.Returning)/float64(report.New+report.Returning)*100)
	report.DataPoints = r.getDataPoints(start, end, ctx)

	pastStart, pastEnd := r.getPastTimes(start, end)

	report.DataPointsPast = r.getDataPoints(pastStart, pastEnd, ctx)

	return report
}

func (r *ReportManager) countVisitors() int {
	visitors := map[uint]bool{}
	for _, v := range r.Visits {
		visitors[v.VisitorID] = true
	}
	return len(visitors)
}

func (r *ReportManager) countBouncedVisits() int {
	count := 0

	for _, v := range r.Visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		if cnt == 1 {
			count++
		}
	}

	return count
}

func (r *ReportManager) getBounceRate() float64 {
	visits := len(r.Visits)
	if visits > 0 {
		blounceRate := float64(r.countBouncedVisits()) / float64(visits) * float64(100)
		return blounceRate
	}
	return 0
}

func (r *ReportManager) countNew() int {
	return r.countVisitors()
}

func (r *ReportManager) countReturning() int {
	newCount := r.countNew()
	visits := len(r.Visits)
	return visits - newCount
}

func (r *ReportManager) getDataPoints(start, end time.Time, ctx *iris.Context) []int {
	var visits []*Visit
	var dataPoints []int
	// now := time.Now().UTC()
	// session := ctx.Session()
	// offset := session.Get("offset")
	// offsetInt, err := strconv.Atoi(offset.(string))
	// if err != nil {
	// 	log.Printf("[helpers.go] Error parsing offset: %s", err)
	// }

	// utcNow := now.Add(time.Minute * time.Duration(-offsetInt))
	// utcDayNow := utcNow.Day()
	daysAgo := int(round(time.Since(start).Hours() / 24))

	// log.Println()

	if time.Since(start).Minutes()-time.Since(end).Minutes() < 1440 {
		for i := 0; i < 24; i++ {
			older := start.UTC().Add(time.Duration(i) * time.Hour)
			newer := start.UTC().Add(time.Duration(i+1) * time.Hour).Add(-time.Microsecond)
			// log.Println(older)
			// log.Println(newer)
			if i == 0 {
				visits = r.Website.getVisits(&older, &newer)
			} else {
				visits = r.Website.getVisitsPrecise(&older, &newer)
			}
			dataPoints = append(dataPoints, len(visits))
		}
	} else {
		first := true
		for ; daysAgo > 0; daysAgo-- {
			if first {
				visits = r.Website.getVisits(getTimeDaysAgo(daysAgo, ctx), getTimeDaysAgo(daysAgo-1, ctx))
			} else {
				visits = r.Website.getVisitsPrecise(getTimeDaysAgo(daysAgo, ctx), getTimeDaysAgo(daysAgo-1, ctx))
			}
			dataPoints = append(dataPoints, len(visits))
			first = false
		}
	}
	return dataPoints
}

func (r *ReportManager) getTimePerVisit() string {
	seconds := 0

	for _, v := range r.Visits {
		var pvs []*PageView
		db.Order("id").Where(&PageView{VisitID: v.ID}).Find(&pvs)
		if len(pvs) > 1 {
			sinceOlder := time.Since(pvs[0].CreatedAt)
			sinceNewer := time.Since(pvs[len(pvs)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	var d time.Duration

	if len(r.Visits) > 0 {
		d = time.Duration(time.Second * time.Duration(seconds/len(r.Visits)))
	} else {
		d = time.Duration(0)
	}

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (r *ReportManager) getTimeAllVisits() string {
	seconds := 0

	for _, v := range r.Visits {
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

func (r *ReportManager) getPageViewsPerVisit() string {
	count := 0

	for _, v := range r.Visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		count += cnt
	}

	return fmt.Sprintf("%.2f", float64(count)/float64(len(r.Visits)))
}

func (r *ReportManager) getDateRangeType(start, end time.Time, ctx *iris.Context) string {
	dateRangeType := "Date Range"
	now := time.Now().UTC()
	session := ctx.Session()
	offset := session.Get("offset")
	offsetInt, err := strconv.Atoi(offset.(string))
	if err != nil {
		log.Printf("[helpers.go] Error parsing offset: %s", err)
	}

	utcDayNow := now.Add(time.Minute * time.Duration(-offsetInt)).Day()

	if utcDayNow == start.AddDate(0, 0, 1).UTC().Day() && utcDayNow == end.UTC().Day() {
		dateRangeType = "Today"
	} else if utcDayNow == start.AddDate(0, 0, 2).UTC().Day() && utcDayNow == end.AddDate(0, 0, 1).UTC().Day() {
		dateRangeType = "Yesterday"
	} else if utcDayNow == start.AddDate(0, 0, 7).UTC().Day() && utcDayNow == end.UTC().Day() {
		dateRangeType = "Last 7 Days"
	} else if utcDayNow == start.AddDate(0, 0, 30).UTC().Day() && utcDayNow == end.UTC().Day() {
		dateRangeType = "Last 30 Days"
	} else if utcDayNow == start.UTC().AddDate(0, 0, end.UTC().Day()).Day() {
		dateRangeType = "This Month"
	} else {
		dateRangeType = "Last Month"
	}

	return dateRangeType
}

func (r *ReportManager) getChartScale(start, end time.Time, ctx *iris.Context) []string {
	chartScale := []string{}
	utcNow := time.Now().UTC()
	utcDayNow := utcNow.Day()

	if utcDayNow == start.AddDate(0, 0, 1).UTC().Day() && utcDayNow == end.UTC().Day() || utcDayNow == start.AddDate(0, 0, 2).UTC().Day() && utcDayNow == end.AddDate(0, 0, 1).UTC().Day() {
		chartScale = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	} else if utcDayNow == start.AddDate(0, 0, 7).UTC().Day() && utcDayNow == end.UTC().Day() {
		for i := -6; i <= 0; i++ {
			item := utcNow.AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(utcNow.AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	} else if utcDayNow == start.AddDate(0, 0, 30).UTC().Day() && utcDayNow == end.UTC().Day() {
		for i := -29; i <= 0; i++ {
			item := time.Now().AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(time.Now().AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	} else if utcDayNow == start.UTC().AddDate(0, 0, end.UTC().Day()).Day() {
		timeCounter := utcNow.AddDate(0, 0, -utcNow.Day()+1)
		for timeCounter.Month() == utcNow.Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	} else {
		timeCounterMonth := utcNow.AddDate(0, -1, 0)
		timeCounter := timeCounterMonth.AddDate(0, 0, -timeCounterMonth.Day()+1)
		for timeCounter.Month() == timeCounterMonth.Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	}

	return chartScale
}

func (r *ReportManager) getPastTimes(start, end time.Time) (time.Time, time.Time) {
	utcNow := time.Now().UTC()
	utcDayNow := utcNow.Day()

	if utcDayNow == start.AddDate(0, 0, 1).UTC().Day() && utcDayNow == end.UTC().Day() || utcDayNow == start.AddDate(0, 0, 2).UTC().Day() && utcDayNow == end.AddDate(0, 0, 1).UTC().Day() {
		start = start.AddDate(0, 0, -1)
		end = end.AddDate(0, 0, -1)
	} else if utcDayNow == start.AddDate(0, 0, 7).UTC().Day() && utcDayNow == end.UTC().Day() {
		start = start.AddDate(0, 0, -7)
		end = end.AddDate(0, 0, -7)
	} else if utcDayNow == start.AddDate(0, 0, 30).UTC().Day() && utcDayNow == end.UTC().Day() {
		start = start.AddDate(0, 0, -30)
		end = end.AddDate(0, 0, -30)
	} else if utcDayNow == start.UTC().AddDate(0, 0, end.UTC().Day()).Day() {

	} else {

	}

	return start, end
}

func initReport() {
	rm = &ReportManager{}
}
