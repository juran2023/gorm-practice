package main

import (
	"log"
	"os"
	"time"

	"encoding/json"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model
	Name      string
	Age       int
	Languages []Language `gorm:"many2many:UserLanguage;"`
}

type Language struct {
	gorm.Model
	Name  string
	Users []User `gorm:"many2many:UserLanguage;"`
}

type UserLanguage struct {
	UserID     uint `gorm:"primaryKey"`
	LanguageID uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
			Colorful: true,
		},
	)

	dsn := "host=localhost user=postgres password=123456 dbname=dvdrental sslmode=disable timezone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{}, &Language{}, &UserLanguage{})

	var users []User
	var languages []Language

	db.Preload("Languages").Find(&users)

	var byteArray []byte

	byteArray, err = json.MarshalIndent(users, "", "  ")
	if err != nil {
		panic("failed to marshal users: " + err.Error())
	}

	fmt.Println(string(byteArray))

	db.Preload("Users").Find(&languages)

	byteArray, err = json.MarshalIndent(languages, "", "  ")
	if err != nil {
		panic("failed to marshal languages: " + err.Error())
	}

	fmt.Println(string(byteArray))
}
