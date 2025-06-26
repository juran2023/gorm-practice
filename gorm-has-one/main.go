package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

type User struct {
	gorm.Model
	Name       string
	Email      string
	Role       string
	CreditCard CreditCard
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID int
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

	db.AutoMigrate(&User{}, &CreditCard{})

	// Create a new user
	user := User{
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}
	db.Create(&user)

	// Create a credit card for the user
	creditCard := CreditCard{
		Number: "1234-5678-9012-3456",
		UserID: int(user.ID),
	}
	db.Create(&creditCard)

	// Query the user and preload the credit card
	var fetchedUser User
	db.Preload("CreditCard").First(&fetchedUser, user.ID)

	// Marshal to JSON and print
	byteArr, err := json.MarshalIndent(fetchedUser, "", "  ")
	if err != nil {
		panic("json.Marshal error")
	}
	fmt.Println("Fetched User with Credit Card:\n", string(byteArr))
}
