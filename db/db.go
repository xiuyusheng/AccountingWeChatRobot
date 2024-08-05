package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func init() {
	var err error
	var log logger.Interface
	DB, err = gorm.Open(sqlite.Open("./accountbook.sqlite"), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: log,
	})
	if err != nil {
		panic(err)
	}
}
