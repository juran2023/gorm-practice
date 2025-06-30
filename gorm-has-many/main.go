package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model
	CreditCards []CreditCard `gorm:"foreignKey:UserID" constraint:"OnUpdate:CASCADE,OnDelete:SET NULL"`
	Name        string
	ManagerID   uint
	Team        []User `gorm:"foreignKey:ManagerID" constraint:"OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint
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

	// // Create a user
	// user := User{Name: "jinzhu"}
	// db.Create(&user)

	// // Create multiple credit cards for the user
	// creditCard1 := CreditCard{Number: "1111-2222-3333-4444", UserID: user.ID}
	// creditCard2 := CreditCard{Number: "5555-6666-7777-8888", UserID: user.ID}
	// db.Create(&creditCard1)
	// db.Create(&creditCard2)

	// // Query the user and preload credit cards
	// var fetchedUser User
	// db.Preload("CreditCards").First(&fetchedUser, user.ID)

	// // Marshal to JSON and print
	// byteArr, err := json.MarshalIndent(fetchedUser, "", "  ")
	// if err != nil {
	// 	panic("json.Marshal error")
	// }
	// fmt.Println("Fetched User with Credit Cards:\n", string(byteArr))

	// 自引用
	// employee1 := User{
	// 	Name:      "employee1",
	// 	ManagerID: 630,
	// }
	// employee2 := User{
	// 	Name:      "employee2",
	// 	ManagerID: 1,
	// }
	// db.Create(&employee1)
	// db.Create(&employee2)

	// var manager User
	// // 能否Preload CreditCards 和 Team
	// db.Preload("CreditCards").Preload("Team").First(&manager, 630)

	// byteArr, err = json.MarshalIndent(manager, "", "  ")
	// if err != nil {
	// 	panic("json.Marshal error")
	// }
	// fmt.Println("Manager with CreditCards and Team:\n", string(byteArr))

}
