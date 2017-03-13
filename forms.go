package main

type SignInForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

type SignUpForm struct {
	Email     string `form:"email"`
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
}
