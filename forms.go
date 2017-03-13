package main

type SignInForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}
