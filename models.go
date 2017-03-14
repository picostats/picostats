package main

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email    string `sql:"size:255" unique_index`
	Password string `sql:"size:255"`
	Verified bool   `sql:"not null"`
}

type Website struct {
	gorm.Model
	Owner   *User
	OwnerID int `sql:"index"`
}

type PageView struct {
	gorm.Model
	Website   *Website
	WebsiteID int `sql:"index"`
}
