package main

import (
	"github.com/jinzhu/gorm"
)

// func GenerateAnonymousUser() sessionauth.User {
// 	return &User{}
// }

// User struct represents user model
type User struct {
	gorm.Model
	Email         string `sql:"size:255" unique_index`
	Password      string `sql:"size:255"`
	Verified      bool   `sql:"not null"`
	authenticated bool
}
