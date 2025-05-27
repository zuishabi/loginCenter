package database

import "time"

type UserInfo struct {
	UID       uint32 `gorm:"primarykey"`
	CreatedAt time.Time
	Password  string `gorm:"Index:idx_email_psw"`
	UserName  string `gorm:"uniqueIndex;size:10"`
	UserEmail string `gorm:"Index:idx_email_psw"`
}

func Update() {
	Db.AutoMigrate(&UserInfo{})
}
