package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model
	Name        string
	Age         int `gorm:"check:chk_users_age, age % 10 <> 0"` // 过9不过10
	CreditCards []CreditCard `gorm:"foreignKey:UserID" constraint:"OnUpdate:CASCADE,OnDelete:SET NULL"`
	Languages   []Language   `gorm:"many2many:UserLanguage;"`
}

type CreditCard struct {
	gorm.Model
	Number     string
	UserID     uint
	ExpireDate time.Time
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
		SkipDefaultTransaction: true,
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{}, &Language{}, &UserLanguage{})

	user := User{
		Name: "old man",
		Age: 100,
	}
	db.Migrator().DropConstraint(&User{}, "chk_users_age")

	if err = db.Model(&User{}).Create(&user).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
	}

}
