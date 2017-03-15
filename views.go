package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	pd := newPageData(ctx)
	if isSignedIn(ctx) {
		pd.User.redirectToDefaultWebsite(ctx)
	}
	pd.Form = SignInForm{}
	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func signInPostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	sif := &SignInForm{}
	err := ctx.ReadForm(sif)
	if err != nil {
		log.Printf("[views.go] Error reading SignInForm: %s", err)
	}

	user := &User{}
	db.Where("email = ?", sif.Email).First(user)

	if user.ID != 0 {
		if user.Password == getMD5Hash(sif.Password) {
			signIn(ctx, user)
			pd := newPageData(ctx)
			pd.User.redirectToDefaultWebsite(ctx)
		} else {
			err := errors.New("Email or password is wrong, please try again.")
			pd.Errors = append(pd.Errors, &err)
		}
	} else {
		err := errors.New("Email or password is wrong, please try again.")
		pd.Errors = append(pd.Errors, &err)
	}

	pd.Form = &sif

	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func signOutView(ctx *iris.Context) {
	signOut(ctx)
	ctx.Redirect(conf.AppUrl + APP_PATH + "/sign-in")
}

func signUpView(ctx *iris.Context) {
	pd := newPageData(ctx)
	if isSignedIn(ctx) {
		pd.User.redirectToDefaultWebsite(ctx)
	}
	pd.Form = SignUpForm{}
	ctx.Render("sign-up.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func signUpPostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	suf := &SignUpForm{}
	err := ctx.ReadForm(suf)
	if err != nil {
		log.Printf("[views.go] Error reading SignUpForm: %s", err)
	}

	user := &User{}
	db.Where("email = ?", suf.Email).First(user)

	if user.ID == 0 {
		if suf.Password1 == suf.Password2 {
			user := &User{Email: suf.Email, Password: getMD5Hash(suf.Password1)}
			db.Create(user)
			pd.User.redirectToDefaultWebsite(ctx)
		} else {
			err := errors.New("Passwords don't match, please try again.")
			pd.Errors = append(pd.Errors, &err)
		}

	} else {
		err := errors.New("User with this email address already exists.")
		pd.Errors = append(pd.Errors, &err)
	}

	pd.Form = &suf

	ctx.Render("sign-up.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func collectImgView(ctx *iris.Context) {
	pv := &PageViewRequest{
		WebsiteID:  ctx.URLParam("w"),
		Path:       ctx.URLParam("p"),
		Hostname:   ctx.URLParam("h"),
		Title:      ctx.URLParam("t"),
		Language:   ctx.URLParam("l"),
		Resolution: ctx.URLParam("s"),
		Referrer:   ctx.URLParam("r"),
		IpAddress:  strings.Split(ctx.Request.RemoteAddr, ":")[0],
	}

	pvJson, err := json.Marshal(pv)
	if err != nil {
		log.Printf("[views.go] Error in tracking image: %s", err)
	}

	red.LPush("pvs", string(pvJson))

	bytes := getTrackerImageBytes()
	ctx.SetContentType("image/png")
	ctx.Write(bytes)
}

func accountView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("account.html", pd)
}

func newWebsiteView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("websites-new.html", pd)
}

func newWebsitePostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	wf := &WebsiteForm{}
	err := ctx.ReadForm(wf)
	if err != nil {
		log.Printf("[views.go] Error reading WebsiteForm: %s", err)
	}

	w := &Website{
		OwnerID: pd.User.ID,
		Name:    wf.Name,
		Url:     wf.Url,
	}

	db.Create(w)

	w.TrackingCode = aesEncrypt(strconv.Itoa(int(w.ID)))
	db.Save(w)

	ctx.Redirect(conf.AppUrl + APP_PATH + "/websites/" + strconv.Itoa(int(w.ID)))
}

func editWebsiteView(ctx *iris.Context) {
	pd := newPageData(ctx)
	wId, err := ctx.ParamInt64("id")
	if err != nil {
		log.Printf("[views.go] Error getting website id param: %s", err)
	}
	w := &Website{}
	db.First(w, wId)

	if w.OwnerID == pd.User.ID {
		wf := &WebsiteForm{
			Id:      w.ID,
			Name:    w.Name,
			Url:     w.Url,
			Default: w.Default,
		}
		pd.Form = wf
		pd.WebsiteId = w.TrackingCode
		pd.TrackerUrl = strings.Replace(strings.Replace(conf.AppUrl, "https://", "//", -1), "http://", "//", -1) + "/public/tracker.js"
		ctx.Render("websites-edit.html", pd)
	} else {
		session := ctx.Session()
		session.SetFlash("error", "You are not the owner of this website.")
		pd.User.redirectToDefaultWebsite(ctx)
	}
}

func editWebsitePostView(ctx *iris.Context) {
	pd := newPageData(ctx)
	wId, err := ctx.ParamInt64("id")
	if err != nil {
		log.Printf("[views.go] Error getting website id param: %s", err)
	}
	w := &Website{}
	db.First(w, wId)
	session := ctx.Session()
	if w.OwnerID == pd.User.ID {
		wf := &WebsiteForm{}
		err = ctx.ReadForm(wf)
		if err != nil {
			log.Printf("[views.go] Error reading WebsiteForm: %s", err)
		}
		w.Name = wf.Name
		w.Url = wf.Url
		db.Save(w)
		session.SetFlash("success", "Website successfully updated.")
		ctx.Redirect(conf.AppUrl + APP_PATH + "/websites/" + strconv.Itoa(int(w.ID)))
	} else {
		session.SetFlash("error", "You are not the owner of this website.")
		pd.User.redirectToDefaultWebsite(ctx)
	}
}

func websiteMakeDefaultView(ctx *iris.Context) {
	pd := newPageData(ctx)
	wId, err := ctx.ParamInt64("id")
	if err != nil {
		log.Printf("[views.go] Error getting website id param: %s", err)
	}
	w := &Website{}
	db.First(w, wId)
	session := ctx.Session()
	if w.OwnerID == pd.User.ID {
		oldDefault := pd.User.getDefaultWebsite()
		oldDefault.Default = false
		db.Save(oldDefault)
		w.Default = true
		db.Save(w)
		session.SetFlash("success", "You changed the default website.")
		pd.User.redirectToDefaultWebsite(ctx)
	} else {
		session.SetFlash("error", "You are not the owner of this website.")
		pd.User.redirectToDefaultWebsite(ctx)
	}
}

func websiteView(ctx *iris.Context) {
	pd := newPageData(ctx)
	wId, err := ctx.ParamInt64("id")
	if err != nil {
		log.Printf("[views.go] Error getting website id param: %s", err)
	}
	w := &Website{}
	db.First(w, wId)

	if w.OwnerID == pd.User.ID {
		pd.Form = w

		session := ctx.Session()
		startStr := session.GetString("date-range-start")
		endStr := session.GetString("date-range-end")

		if len(startStr) == 0 {
			t := getTimeDaysAgo(7)
			startStr = strconv.Itoa(int(t.Unix()))
		}
		if len(endStr) == 0 {
			t := getTimeDaysAgo(0)
			endStr = strconv.Itoa(int(t.Unix()))
		}

		startInt, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			log.Printf("[views.go] Error parsing timestamp: %s", err)
		}
		start := time.Unix(startInt, 0)

		endInt, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			log.Printf("[views.go] Error parsing timestamp: %s", err)
		}
		end := time.Unix(endInt, 0)

		pd.DataRangeStartSubtract = int(time.Since(start).Hours() / 24)
		if time.Since(end).Hours() > 0 {
			pd.DataRangeEndSubract = int(time.Since(end).Hours()/24) + 1
		} else {
			pd.DataRangeEndSubract = 0
		}

		pd.DateRangeType = getDateRangeType(pd.DataRangeStartSubtract, pd.DataRangeEndSubract)
		pd.ChartScale = getChartScale(pd.DataRangeStartSubtract, pd.DataRangeEndSubract)

		var dataPoints []int
		var dataPointsPast []int

		if (pd.DataRangeStartSubtract == 0 && pd.DataRangeEndSubract == 0) || (pd.DataRangeStartSubtract == 1 && pd.DataRangeEndSubract == 1) {
			dataPoints = w.getDataPointsHourly(pd.DataRangeStartSubtract)
			dataPointsPast = w.getDataPointsHourly(pd.DataRangeStartSubtract + 1)
		} else {
			dataPoints = w.getDataPoints(pd.DataRangeStartSubtract+1, pd.DataRangeStartSubtract+1)
			dataPointsPast = w.getDataPoints((pd.DataRangeStartSubtract+1)*2, pd.DataRangeStartSubtract+1)
		}

		pd.Report = &Report{
			PageViews:         w.countPageViews(&start, &end),
			Users:             w.countUsers(&start, &end),
			Visits:            w.countVisits(&start, &end),
			New:               w.countNew(&start, &end),
			Returning:         w.countReturning(&start, &end),
			DataPoints:        dataPoints,
			DataPointsPast:    dataPointsPast,
			BounceRate:        fmt.Sprintf("%.2f", w.getBounceRate(&start, &end)),
			TimePerVisit:      w.getTimePerVisit(&start, &end),
			TimeTotal:         w.getTimeAllVisits(&start, &end),
			PageViewsPerVisit: w.getPageViewsPerVisit(&start, &end),
		}
		pd.Report.NewPercentage = fmt.Sprintf("%.2f", float64(pd.Report.New)/float64(pd.Report.New+pd.Report.Returning)*100)
		pd.Report.ReturningPercentage = fmt.Sprintf("%.2f", float64(pd.Report.Returning)/float64(pd.Report.New+pd.Report.Returning)*100)
		ctx.Render("website.html", pd)
	} else {
		session := ctx.Session()
		session.SetFlash("error", "You are not the owner of this website.")
		pd.User.redirectToDefaultWebsite(ctx)
	}
}

func changeDateRangeView(ctx *iris.Context) {
	wId := ctx.Param("id")

	drf := &DateRangeForm{}
	err := ctx.ReadForm(drf)
	if err != nil {
		log.Printf("[views.go] Error reading DateRangeForm: %s", err)
	}

	session := ctx.Session()
	session.Set("date-range-start", drf.Start)
	session.Set("date-range-end", drf.End)

	ctx.Redirect(conf.AppUrl + APP_PATH + "/" + wId)
}

func websiteDeleteView(ctx *iris.Context) {
	pd := newPageData(ctx)
	wId, err := ctx.ParamInt64("id")
	if err != nil {
		log.Printf("[views.go] Error getting website id param: %s", err)
	}
	w := &Website{}
	db.First(w, wId)
	session := ctx.Session()
	if w.OwnerID == pd.User.ID {
		session.SetFlash("success", "Website successfully deleted.")
		db.Delete(w)
	} else {
		session.SetFlash("error", "You are not the owner of this website.")
	}
	pd.User.redirectToDefaultWebsite(ctx)
}
