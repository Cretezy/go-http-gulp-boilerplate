package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string
	Password string
}

type Session struct {
	gorm.Model
	User         User
	UserID       uint
	Token        string
	LoginTime    time.Time
	LastSeenTime time.Time
}
