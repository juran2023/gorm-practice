package main

import (
	"encoding/json"
	"fmt"
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
	Age         int
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
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{}, &Language{}, &UserLanguage{})

	// Preload With Conditions
	var user User
	db.Model(&User{}).Where("id = ?", 11).Preload("CreditCards", "number NOT LIKE ?", "%2%").First(&user)

	byteArray, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		panic("failed to marshal user: " + err.Error())
	}

	fmt.Println(string(byteArray))

	// Custom Preloading SQL
	db.Preload("CreditCards", func(db *gorm.DB) *gorm.DB {
		return db.Order("credit_cards.number ASC")
	}).First(&user)

	byteArray, err = json.MarshalIndent(user, "", "  ")
	if err != nil {
		panic("failed to marshal user: " + err.Error())
	}

	fmt.Println(string(byteArray))

	//Nested Preloading
	// db.Preload("Orders", "state = ?", "paid").Preload("Orders.OrderItems").Find(&users)

	//Embedded Preloading

}
