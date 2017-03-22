package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
)

func initDB() *gorm.DB {
	log.Printf("[db.go] pg: %s", conf.DBUrl)
	db, err := gorm.Open(conf.DBType, conf.DBUrl)

	counter := 0

	for err != nil && counter < 60 {
		log.Printf("[db.go] error: %s", err)
		time.Sleep(time.Second)
		db, err = gorm.Open(conf.DBType, conf.DBUrl)
		counter++
	}

	db.DB()
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(conf.LogSQL)

	db.AutoMigrate(&User{}, &Website{}, &Visitor{}, &Page{}, &PageView{}, &Visit{}, &ReportModel{})

	return db
}
