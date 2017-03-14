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
	OwnerID int    `sql:"index"`
	Name    string `sql:"size:255"`
	Url     string `sql:"size:255"`
}

type Visitor struct {
	gorm.Model
	IpAddress  string `sql:"size:255"`
	Resolution string `sql:"size:255"`
	Language   string `sql:"size:255"`
}

type Page struct {
	gorm.Model
	Hostname string `sql:"size:255"`
	Path     string `sql:"size:255"`
	Title    string `sql:"size:255"`
}

type PageView struct {
	gorm.Model
	Website   *Website
	WebsiteID int `sql:"index"`
	Visitor   *Visitor
	VisitorID int `sql:"index"`
	Page      *Page
	PageID    int `sql:"index"`
}
