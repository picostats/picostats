package main

type SignInForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
	Next     string `form:"next"`
}

type SignUpForm struct {
	Email     string `form:"email"`
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
}

type WebsiteForm struct {
	Id      uint   `form:"id"`
	Name    string `form:"name"`
	Url     string `form:"url"`
	Default bool   `form:"default"`
}

type DateRangeForm struct {
	Start string `form:"start"`
	End   string `form:"end"`
}

type PasswordForm struct {
	Password1       string `form:"password1"`
	Password2       string `form:"password2"`
	CurrentPassword string `form:"password"`
}
