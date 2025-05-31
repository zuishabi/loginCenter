package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// mysql-login:3306
var dsn = "root:861214959@tcp(mysql-login:3306)/game?charset=utf8mb4&parseTime=True&loc=Local"
var Db *gorm.DB

func Start() error {
	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}
