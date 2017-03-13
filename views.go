package main

import (
	"errors"
	"log"

	"gopkg.in/kataras/iris.v6"
)

func signInView(ctx *iris.Context) {
	if isSignedIn(ctx) {
		ctx.Redirect(conf.AppUrl + APP_PATH)
		return
	}
	pd := newPageData(ctx)
	pd.Form = SignInForm{}
	ctx.Render("sign-in.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func signInPostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	sif := &SignInForm{}
	err := ctx.ReadForm(sif)
	if err != nil {
		log.Println("[views.go] Error reading SignInForm: %s", err)
	}

	user := &User{}
	db.Where("email = ?", sif.Email).First(user)

	if user.ID != 0 {
		if user.Password == getMD5Hash(sif.Password) {
			signIn(ctx, user)
			ctx.Redirect(conf.AppUrl + APP_PATH)
			return
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
	if isSignedIn(ctx) {
		ctx.Redirect(conf.AppUrl + APP_PATH)
		return
	}
	pd := newPageData(ctx)
	pd.Form = SignUpForm{}
	ctx.Render("sign-up.html", pd, iris.RenderOptions{"layout": "layout2.html"})
}

func signUpPostView(ctx *iris.Context) {
	pd := newPageData(ctx)

	suf := &SignUpForm{}
	err := ctx.ReadForm(suf)
	if err != nil {
		log.Println("[views.go] Error reading SignUpForm: %s", err)
	}

	user := &User{}
	db.Where("email = ?", suf.Email).First(user)

	if user.ID == 0 {
		if suf.Password1 == suf.Password2 {
			user := &User{Email: suf.Email, Password: getMD5Hash(suf.Password1)}
			db.Create(user)
			ctx.Redirect(conf.AppUrl + APP_PATH)
			return
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

func dashboardView(ctx *iris.Context) {
	pd := newPageData(ctx)
	ctx.Render("dashboard.html", pd)
}
