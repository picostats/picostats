package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
}

func (rm *ReportManager) getReport(ctx *iris.Context, w *Website, pd *PageData) *Report {
	reportType := rm.getReportType(ctx)
	session := ctx.Session()

	startStr := session.GetString("date-range-start")
	endStr := session.GetString("date-range-end")

	if len(startStr) == 0 || len(endStr) == 0 {
		startStr, endStr = rm.getDefaultTimesStr(pd.User.TimeOffset)
	}

	startInt, err := strconv.Atoi(startStr)
	if err != nil {
		log.Printf("[report.go] Error parsing timestamp: %s", err)
	}

	endInt, err := strconv.Atoi(endStr)
	if err != nil {
		log.Printf("[report.go] Error parsing timestamp: %s", err)
	}

	repMod := &ReportModel{}
	db.Where("website_id = ? AND start_int = ? AND end_int = ? AND type = ?", w.ID, startInt, endInt, reportType).First(repMod)

	if repMod.ID == 0 {
		return rm.generateNew(reportType, startInt, endInt, w)
	} else {
		report := &Report{
			Visits:              repMod.Visits,
			Visitors:            repMod.Visitors,
			PageViews:           repMod.PageViews,
			BounceRate:          repMod.BounceRate,
			New:                 repMod.New,
			Returning:           repMod.Returning,
			DataPoints:          splitDataPoints(repMod.DataPoints),
			DataPointsPast:      splitDataPoints(repMod.DataPointsPast),
			TimePerVisit:        repMod.TimePerVisit,
			TimeTotal:           repMod.TimeTotal,
			PageViewsPerVisit:   repMod.PageViewsPerVisit,
			NewPercentage:       repMod.NewPercentage,
			ReturningPercentage: repMod.ReturningPercentage,
			DateRangeType:       repMod.DateRangeType,
			ChartScale:          strings.Split(repMod.ChartScale, "|"),
			StartInt:            repMod.StartInt,
			EndInt:              repMod.EndInt,
		}

		return report
	}
}

func (rm *ReportManager) getReportType(ctx *iris.Context) int {
	session := ctx.Session()
	typeStr := session.GetString("date-range-type")

	if len(typeStr) == 0 {
		return REPORT_TYPE_TODAY
	}

	typeInt, err := strconv.Atoi(typeStr)
	if err != nil {
		log.Printf("[report.go] Error parsing typeStr: %s", err)
	}

	return typeInt
}

func (rm *ReportManager) getDefaultTimesStr(offset float64) (string, string) {
	start := time.Now().In(time.UTC).Add(time.Minute * time.Duration(-offset)).Truncate(24 * time.Hour).Add(time.Minute * time.Duration(offset))
	end := start.AddDate(0, 0, 1).Add(-time.Millisecond)
	return strconv.Itoa(int(start.Unix())), strconv.Itoa(int(end.Unix()))
}

func (rm *ReportManager) getDefaultTimes(offset float64, reportType int) (time.Time, time.Time) {
	var start, end time.Time

	switch reportType {
	case REPORT_TYPE_TODAY:
		start = time.Now().In(time.UTC).Add(time.Minute * time.Duration(-offset)).Truncate(24 * time.Hour).Add(time.Minute * time.Duration(offset))
		end = start.AddDate(0, 0, 1).Add(-time.Millisecond)
	case REPORT_TYPE_YESTERDAY:
		start = time.Now().In(time.UTC).Add(time.Minute*time.Duration(-offset)).Truncate(24*time.Hour).Add(time.Minute*time.Duration(offset)).AddDate(0, 0, -1)
		end = start.AddDate(0, 0, 1).Add(-time.Millisecond)
	case REPORT_TYPE_7_DAYS:
		start = time.Now().In(time.UTC).Add(time.Minute*time.Duration(-offset)).Truncate(24*time.Hour).Add(time.Minute*time.Duration(offset)).AddDate(0, 0, -6)
		end = start.AddDate(0, 0, 7).Add(-time.Millisecond)
	case REPORT_TYPE_30_DAYS:
		start = time.Now().In(time.UTC).Add(time.Minute*time.Duration(-offset)).Truncate(24*time.Hour).Add(time.Minute*time.Duration(offset)).AddDate(0, 0, -29)
		end = start.AddDate(0, 0, 30).Add(-time.Millisecond)
	case REPORT_TYPE_THIS_MONTH:
		end = time.Now().In(time.UTC).Add(time.Minute*time.Duration(-offset)).Truncate(24*time.Hour).Add(time.Minute*time.Duration(offset)).AddDate(0, 0, 1)
		start = end.AddDate(0, 0, -end.Day())
		end = end.Add(-time.Microsecond)
	case REPORT_TYPE_LAST_MONTH:
		start = time.Now().In(time.UTC).Add(time.Minute*time.Duration(-offset)).Truncate(24*time.Hour).Add(time.Minute*time.Duration(offset)).AddDate(0, -1, 0)
		start = start.AddDate(0, 0, -start.Day())
		end = time.Now().In(time.UTC).Truncate(24 * time.Hour).Add(time.Minute * time.Duration(offset))
		end = end.AddDate(0, 0, -end.Day()).Add(-time.Microsecond)
	default:
		start = time.Now().In(time.UTC).Truncate(24 * time.Hour).Add(time.Minute * time.Duration(offset))
		end = start.AddDate(0, 0, 1).Add(-time.Microsecond)
	}

	return start, end
}

func (rm *ReportManager) generateNew(reportType, startInt, endInt int, w *Website) *Report {
	r := &Report{StartInt: int64(startInt), EndInt: int64(endInt)}
	rh := &ReportHolder{Website: w, Type: reportType, Report: r}

	rh.generateReport()

	// repMod := &ReportModel{
	// 	WebsiteID:     w.ID,
	// 	StartInt:      rh.Report.StartInt,
	// 	EndInt:        rh.Report.EndInt,
	// 	Type:          reportType,
	// 	DateRangeType: rh.Report.DateRangeType,
	// 	ChartScale:    strings.Join(rh.Report.ChartScale, "|"),

	// 	Visits:              rh.Report.Visits,
	// 	Visitors:            rh.Report.Visitors,
	// 	PageViews:           rh.Report.PageViews,
	// 	BounceRate:          rh.Report.BounceRate,
	// 	New:                 rh.Report.New,
	// 	Returning:           rh.Report.Returning,
	// 	DataPoints:          joinDataPoints(rh.Report.DataPoints),
	// 	DataPointsPast:      joinDataPoints(rh.Report.DataPointsPast),
	// 	TimePerVisit:        rh.Report.TimePerVisit,
	// 	TimeTotal:           rh.Report.TimeTotal,
	// 	PageViewsPerVisit:   rh.Report.PageViewsPerVisit,
	// 	NewPercentage:       rh.Report.NewPercentage,
	// 	ReturningPercentage: rh.Report.ReturningPercentage,
	// }

	// db.Create(repMod)

	return rh.Report
}

type ReportHolder struct {
	Type          int
	PageViews     []*PageView
	Visits        []*Visit
	VisitsPrecise []*Visit
	Website       *Website
	Report        *Report
}

func (rh *ReportHolder) generateReport() *Report {
	start := time.Unix(rh.Report.StartInt, 0)
	end := time.Unix(rh.Report.EndInt, 0)

	rh.PageViews = rh.Website.getPageViews(&start, &end)
	rh.Visits = rh.Website.getVisits(&start, &end)
	rh.VisitsPrecise = rh.Website.getVisitsPrecise(&start, &end)

	rh.Report.DateRangeType = rh.getDateRangeType()
	rh.Report.ChartScale = rh.getChartScale()

	rh.Report.PageViews = len(rh.PageViews)
	rh.Report.Visitors = rh.countVisitors()
	rh.Report.Visits = len(rh.Visits)
	rh.Report.New = rh.countNew()
	rh.Report.Returning = rh.countReturning()
	rh.Report.BounceRate = fmt.Sprintf("%.2f", rh.getBounceRate())
	rh.Report.TimePerVisit = rh.getTimePerVisit()
	rh.Report.TimeTotal = rh.getTimeAllVisits()
	rh.Report.PageViewsPerVisit = rh.getPageViewsPerVisit()
	rh.Report.NewPercentage = fmt.Sprintf("%.2f", float64(rh.Report.New)/float64(rh.Report.New+rh.Report.Returning)*100)
	rh.Report.ReturningPercentage = fmt.Sprintf("%.2f", float64(rh.Report.Returning)/float64(rh.Report.New+rh.Report.Returning)*100)
	rh.Report.DataPoints = rh.getDataPoints(start, end)

	pastStart, pastEnd := rh.getPastTimes(start, end)
	rh.Report.DataPointsPast = rh.getDataPoints(pastStart, pastEnd)

	rh.Report = rh.Report

	return rh.Report
}

func (rh *ReportHolder) countVisitors() int {
	visitors := map[uint]bool{}
	for _, v := range rh.Visits {
		visitors[v.VisitorID] = true
	}
	return len(visitors)
}

func (rh *ReportHolder) countBouncedVisits() int {
	count := 0

	for _, v := range rh.Visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		if cnt == 1 {
			count++
		}
	}

	return count
}

func (rh *ReportHolder) getBounceRate() float64 {
	visits := len(rh.Visits)
	if visits > 0 {
		blounceRate := float64(rh.countBouncedVisits()) / float64(visits) * float64(100)
		return blounceRate
	}
	return 0
}

func (rh *ReportHolder) countNew() int {
	return rh.countVisitors()
}

func (rh *ReportHolder) countReturning() int {
	newCount := rh.countNew()
	visits := len(rh.Visits)
	return visits - newCount
}

func (rh *ReportHolder) getTimePerVisit() string {
	seconds := 0

	for _, v := range rh.Visits {
		var pvs []*PageView
		db.Order("id").Where(&PageView{VisitID: v.ID}).Find(&pvs)
		if len(pvs) > 1 {
			sinceOlder := time.Since(pvs[0].CreatedAt)
			sinceNewer := time.Since(pvs[len(pvs)-1].CreatedAt)
			seconds += int(sinceOlder.Seconds() - sinceNewer.Seconds())
		}
	}

	var d time.Duration

	if len(rh.Visits) > 0 {
		d = time.Duration(time.Second * time.Duration(seconds/len(rh.Visits)))
	} else {
		d = time.Duration(0)
	}

	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

func (rh *ReportHolder) getTimeAllVisits() string {
	seconds := 0

	for _, v := range rh.Visits {
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

func (rh *ReportHolder) getPageViewsPerVisit() string {
	count := 0

	for _, v := range rh.Visits {
		var cnt int
		var pvs []*PageView
		db.Where(&PageView{VisitID: v.ID}).Find(&pvs).Count(&cnt)
		count += cnt
	}

	return fmt.Sprintf("%.2f", float64(count)/float64(len(rh.Visits)))
}

func (rh *ReportHolder) getDateRangeType() string {
	switch rh.Type {
	case REPORT_TYPE_TODAY:
		return "Today"
	case REPORT_TYPE_YESTERDAY:
		return "Yesterday"
	case REPORT_TYPE_7_DAYS:
		return "Last 7 Days"
	case REPORT_TYPE_30_DAYS:
		return "Last 30 Days"
	case REPORT_TYPE_THIS_MONTH:
		return "This Month"
	case REPORT_TYPE_LAST_MONTH:
		return "Last Month"
	default:
		return "Today"
	}
}

func (rh *ReportHolder) getChartScale() []string {
	chartScale := []string{}

	switch rh.Type {
	case REPORT_TYPE_TODAY:
		chartScale = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	case REPORT_TYPE_YESTERDAY:
		chartScale = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	case REPORT_TYPE_7_DAYS:
		for i := -6; i <= 0; i++ {
			item := time.Now().AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(time.Now().AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	case REPORT_TYPE_30_DAYS:
		for i := -29; i <= 0; i++ {
			item := time.Now().AddDate(0, 0, i).Month().String()[0:3] + " " + strconv.Itoa(time.Now().AddDate(0, 0, i).Day())
			chartScale = append(chartScale, item)
		}
	case REPORT_TYPE_THIS_MONTH:
		timeCounter := time.Now().AddDate(0, 0, -time.Now().Day()+1)
		for timeCounter.Month() == time.Now().Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	case REPORT_TYPE_LAST_MONTH:
		timeCounterMonth := time.Now().AddDate(0, -1, 0)
		timeCounter := timeCounterMonth.AddDate(0, 0, -timeCounterMonth.Day()+1)
		for timeCounter.Month() == timeCounterMonth.Month() {
			chartScale = append(chartScale, timeCounter.Month().String()[0:3]+" "+strconv.Itoa(timeCounter.Day()))
			timeCounter = timeCounter.AddDate(0, 0, 1)
		}
	default:
		chartScale = []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	}

	return chartScale
}

func (rh *ReportHolder) getPastTimes(start, end time.Time) (time.Time, time.Time) {
	switch rh.Type {
	case REPORT_TYPE_TODAY:
		start = start.AddDate(0, 0, -1)
		end = end.AddDate(0, 0, -1)
	case REPORT_TYPE_YESTERDAY:
		start = start.AddDate(0, 0, -1)
		end = end.AddDate(0, 0, -1)
	case REPORT_TYPE_7_DAYS:
		start = start.AddDate(0, 0, -7)
		end = end.AddDate(0, 0, -7)
	case REPORT_TYPE_30_DAYS:
		start = start.AddDate(0, 0, -30)
		end = end.AddDate(0, 0, -30)
	case REPORT_TYPE_THIS_MONTH:
		start = start.AddDate(0, -1, 0)
		end = start.AddDate(0, 1, 0).Add(-time.Second)
	case REPORT_TYPE_LAST_MONTH:
		start = start.AddDate(0, -2, 0)
		end = start.AddDate(0, 1, 0).Add(-time.Second)
	default:
		start = start.AddDate(0, 0, -1)
		end = end.AddDate(0, 0, -1)
	}

	return start, end
}

func (rh *ReportHolder) getDataPoints(start, end time.Time) []int {
	var visits []*Visit
	var dataPoints []int

	if time.Since(start).Minutes()-time.Since(end).Minutes() < 1440 {
		for i := 0; i < 24; i++ {
			older := start.Add(time.Duration(i) * time.Hour)
			newer := start.Add(time.Duration(i+1) * time.Hour).Add(-time.Microsecond)
			if i == 0 {
				visits = rh.Website.getVisits(&older, &newer)
			} else {
				visits = rh.Website.getVisitsPrecise(&older, &newer)
			}
			dataPoints = append(dataPoints, len(visits))
		}
	} else {
		first := true
		limit := ((time.Duration(time.Since(start).Minutes()-time.Since(end).Minutes()) + 1) * time.Minute).Hours() / 24
		for i := 0; i < int(limit); i++ {
			older := start.AddDate(0, 0, i)
			newer := older.AddDate(0, 0, 1).Add(-time.Microsecond)

			if first {
				visits = rh.Website.getVisits(&older, &newer)
			} else {
				visits = rh.Website.getVisitsPrecise(&older, &newer)
			}
			dataPoints = append(dataPoints, len(visits))
			first = false
		}
	}
	return dataPoints
}

func initReport() {
	rm = &ReportManager{}
}
