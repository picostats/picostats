package main

import (
	"github.com/jinzhu/gorm"
	"log"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
)

func initDB() *gorm.DB {
	log.Printf("[db.go] pg: %s", conf.DBUrl)
	db, err := gorm.Open(conf.DBType, conf.DBUrl)

	if err != nil {
		log.Printf("[db.go] error: %s", err)
	}

	db.DB()
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(conf.LogSQL)

	db.AutoMigrate(&User{}, &Website{}, &Visitor{}, &Page{}, &PageView{}, &Visit{})

	return db
}
