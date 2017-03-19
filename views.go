package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tomasen/realip"
	"gopkg.in/kataras/iris.v6"
)

func redirectView(ctx *iris.Context) {
	pd := newPageData(ctx)
	if isSignedIn(ctx) {
		pd.User.redirectToDefaultWebsite(ctx)
	} else {
		ctx.Redirect("/")
	}
}

func signInView(ctx *iris.Context) {
	pd := newPageData(ctx)
	pd.TitlePrefix = "Sign In | "
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

	session := ctx.Session()
	session.Set("offset", sif.Offset)

	user := &User{}
	db.Where("email = ?", sif.Email).First(user)

	if user.ID != 0 {
		if user.Password == getMD5Hash(sif.Password) {
			if user.Verified {
				signIn(ctx, user)
				pd := newPageData(ctx)
				if len(sif.Next) > 0 {
					ctx.Redirect(sif.Next)
				} else {
					pd.User.redirectToDefaultWebsite(ctx)
				}
			} else {
				ctx.Redirect(conf.AppUrl + "/verify")
			}
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
	ctx.Redirect(conf.AppUrl + "/sign-in")
}

func signUpView(ctx *iris.Context) {
	pd := newPageData(ctx)
	pd.TitlePrefix = "Sign Up | "
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
			user := &User{Email: suf.Email, Password: getMD5Hash(suf.Password1), MaxWebsites: conf.MaxWebsites}
			db.Create(user)
			verificationLink := conf.AppUrl + "/verify/" + aesEncrypt(strconv.Itoa(int(user.ID)))
			sendVerificationEmail(user.Email, verificationLink)
			ctx.Redirect(conf.AppUrl + "/verify")
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
	ip := realip.RealIP(ctx.Request)

	pv := &PageViewRequest{
		WebsiteID:  ctx.URLParam("w"),
		Path:       ctx.URLParam("p"),
		Hostname:   ctx.URLParam("h"),
		Title:      ctx.URLParam("t"),
		Language:   ctx.URLParam("l"),
		Resolution: ctx.URLParam("s"),
		Referrer:   ctx.URLParam("r"),
		IpAddress:  ip,
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

	if pd.User.countWebsites() >= pd.User.MaxWebsites && pd.User.MaxWebsites != 0 {
		session := ctx.Session()
		session.SetFlash("error", "You've reached maximum number of websites. If you need more, please <a href=\"https://www.picostats.com/pricing/"+strconv.Itoa(int(pd.User.ID))+"\"><strong>purchase</strong></a> PicoStats Premium or install PicoStats on your own server.")
		pd.User.redirectToDefaultWebsite(ctx)
		return
	}

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
		Default: pd.User.countWebsites() == 0,
	}

	db.Create(w)

	w.TrackingCode = getMD5Hash(strconv.Itoa(int(w.ID)))
	db.Save(w)

	ctx.Redirect(conf.AppUrl + "/websites/" + strconv.Itoa(int(w.ID)))
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
		if pd.User.MaxWebsites == 0 || w.Default {
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
			session.SetFlash("error", "You can't access this website when your PicoStats Premium account is inactive.")
			pd.User.redirectToDefaultWebsite(ctx)
		}
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
		ctx.Redirect(conf.AppUrl + "/websites/" + strconv.Itoa(int(w.ID)))
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
		if pd.User.MaxWebsites == 0 || w.Default {
			pd.Form = w

			session := ctx.Session()
			startStr := session.GetString("date-range-start")
			endStr := session.GetString("date-range-end")

			if len(startStr) == 0 {
				t := getTimeDaysAgo(7, ctx)
				startStr = strconv.Itoa(int(t.Unix()))
			}
			if len(endStr) == 0 {
				t := getTimeDaysAgo(0, ctx)
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
				dataPoints = w.getDataPointsHourly(pd.DataRangeStartSubtract, ctx)
				dataPointsPast = w.getDataPointsHourly(pd.DataRangeStartSubtract+1, ctx)
			} else {
				dataPoints = w.getDataPoints(pd.DataRangeStartSubtract+1, pd.DataRangeStartSubtract+1, ctx)
				dataPointsPast = w.getDataPoints((pd.DataRangeStartSubtract+1)*2, pd.DataRangeStartSubtract+1, ctx)
			}

			pd.Report = &Report{
				PageViews:         w.countPageViews(&start, &end),
				Visitors:          w.countVisitors(&start, &end),
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
			session.SetFlash("error", "You can't access this website when your PicoStats Premium account is inactive.")
			pd.User.redirectToDefaultWebsite(ctx)
		}
	} else {
		session := ctx.Session()
		session.SetFlash("error", "You are not the owner of this website.")
		pd.User.redirectToDefaultWebsite(ctx)
	}
}

func changeDateRangeView(ctx *iris.Context) {
	wId := ctx.Param("id")

	drf := &DateRangeForm{}
	// log.Println(drf.Offset)
	err := ctx.ReadForm(drf)
	if err != nil {
		log.Printf("[views.go] Error reading DateRangeForm: %s", err)
	}

	session := ctx.Session()
	session.Set("date-range-start", drf.Start)
	session.Set("date-range-end", drf.End)

	ctx.Redirect(conf.AppUrl + "/" + wId)
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

func changePasswordPost(ctx *iris.Context) {
	pd := newPageData(ctx)

	pf := &PasswordForm{}
	err := ctx.ReadForm(pf)
	if err != nil {
		log.Printf("[views.go] Error reading PasswordForm: %s", err)
	}

	if pf.Password1 == pf.Password2 {
		if pd.User.Password == getMD5Hash(pf.CurrentPassword) {
			pd.User.Password = getMD5Hash(pf.Password1)
			db.Save(pd.User)
			session := ctx.Session()
			session.SetFlash("success", "You have successfully changed your PicoStats password.")
			ctx.Redirect(conf.AppUrl + "/account")
			return
		} else {
			err := errors.New("Your current password is not right, please try again.")
			pd.Errors = append(pd.Errors, &err)
		}
	} else {
		err := errors.New("Passwords are not matching, please try again.")
		pd.Errors = append(pd.Errors, &err)
	}

	ctx.Render("account.html", pd)
}

func accountDeleteView(ctx *iris.Context) {
	pd := newPageData(ctx)
	db.Delete(pd.User)
	ctx.Redirect(conf.AppUrl + "/sign-out")
}

func verifyMessageView(ctx *iris.Context) {
	pd := newPageData(ctx)
	if isSignedIn(ctx) {
		pd.User.redirectToDefaultWebsite(ctx)
	}
	ctx.Render("verify.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func verifyView(ctx *iris.Context) {
	userIdStr := ctx.Param("hash")
	userId, err := strconv.Atoi(aesDecrypt(userIdStr))
	if err != nil {
		log.Printf("[views.go] Atoi err: %s", err)
	} else {
		u := &User{}
		db.First(u, userId)
		u.Verified = true
		db.Save(u)
		signIn(ctx, u)
		session := ctx.Session()
		session.SetFlash("success", "You have successfully verified your email address and activated your PicoStats account.")
		u.redirectToDefaultWebsite(ctx)
	}
}
