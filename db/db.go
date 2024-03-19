package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Sha256   string
	Meta     string
	FileName string
	Mime     string
}

func Connect(sqlitep string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(sqlitep), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Post{})
	return db
}

type Paginator struct {
	DB    *gorm.DB
	Page  int
	Limit int
}

