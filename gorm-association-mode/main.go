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

	// omit
	// user := User{
	// 	Model: gorm.Model{ID: 5},
	// 	Name:  "John",
	// 	Age:   20,
	// 	Languages: []Language{
	// 		{
	// 			Model: gorm.Model{
	// 				ID: 6,
	// 			},
	// 			Name: "Language1",
	// 		},
	// 		{
	// 			Model: gorm.Model{
	// 				ID: 7,
	// 			},
	// 			Name: "Language2",
	// 		},
	// 	},
	// }

	// user1 := User{
	// 	Name: "Johan",
	// 	Age:  20,
	// 	CreditCards: []CreditCard{
	// 		{
	// 			Number:     "121-32-1345",
	// 			ExpireDate: time.Now().AddDate(1, 0, 0),
	// 		},
	// 		{
	// 			Number:     "142-12-1234",
	// 			ExpireDate: time.Now().AddDate(2, 0, 0),
	// 		},
	// 	},
	// }

	// 会维护关联表，但是默认关联对象已在数据库中，所以会报错
	// db.Model(&user).Omit("Languages.*").Create(&user)

	// 完全忽略关联Model和关系表，不会报错
	// db.Model(&user).Omit("Languages").Create(&user)

	var user User
	db.Preload("CreditCards").First(&user, 11)

	// user.Age = 200
	// user.Name = "MakaBaka"
	// user.Model.ID = 0
	// user.CreditCards[0].Number = "1234567890-0000"
	// db.Save(&user)
	// db.Model(&user).Update("age", 22)
	// db.Select("*", "CreditCards.Number").Create(&user)

	// db.Omit("CreditCards.ExpireDate").Create(&user)

	// 删除关联对象
	user = User{Model: gorm.Model{ID: 11}}
	// 1. 不会删除关联对象
	// db.Delete(&user)

	// 2. 会删除关联对象
	// db.Select("CreditCards").Delete(&user)

	// 3. 选择性删除：只删除第一张信用卡
	// var userWithCards User
	// db.Preload("CreditCards").First(&userWithCards, 11)

	// // 确保有信用卡
	// if len(userWithCards.CreditCards) > 0 {
	// 	// 取消关联
	// 	// db.Model(&userWithCards).Association("CreditCards").Delete(userWithCards.CreditCards[0])

	// 	// 只删除第一张信用卡
	// 	db.Delete(&userWithCards.CreditCards[0])
	// }

	// 4. 不会删除关联对象
	// db.Model(&User{}).Select("CreditCards").Where("id = ?", 11).Delete(&User{})

	// Association Mode
	// 1.Finding Associations

	var cards []CreditCard

	db.Model(&user).Association("CreditCards").Find(&cards)

	byteArray, err := json.MarshalIndent(cards, "", "  ")
	if err != nil {
		panic("failed to marshal cards: " + err.Error())
	}

	fmt.Println(string(byteArray))

	// 2.Appending Associations
	// db.Model(&user).Association("CreditCards").Append([]CreditCard{
	// 	{Number: "1234567890"},
	// 	{Number: "1234567891"},
	// })

	// 3. Replacing Associations
	// db.Model(&user).Association("CreditCards").Replace([]CreditCard{
	// 	{Number: "9999999999"},
	// })

	// 4. Deleting Associations
	// db.Model(&user).Association("CreditCards").Delete(cards)

	// 5. Clearing Associations
	// db.Model(&user).Association("CreditCards").Clear()

	// 6. Counting Associations
	count := db.Model(&user).Association("CreditCards").Count()
	fmt.Println(count)

	// 7. Batch Data Handlingvar
	var users = []User{}

	db.Where("id IN ?", []uint{9, 10, 11}).Find(&users)

	fmt.Println("==============================Batch Data Handling==============================")
	// db.Model(&users).Association("CreditCards").Append(&CreditCard{Number: "111111111111111"}, &CreditCard{Number: "666666666"}, &[]CreditCard{{Number: "222222222222222"}, {Number: "333333333333333"}})

	// db.Model(&users).Association("CreditCards").Replace(&CreditCard{Number: "xxxxxxxxxxxxx"}, &CreditCard{Number: "bbbbbbbbbbbbbbbbb"}, &[]CreditCard{{Number: "ccccccccccccc"}, {Number: "ddddddddddddddddd"}})

	// db.Model(&users).Association("CreditCards").Delete(users)

	// db.Model(&users).Association("CreditCards").Clear()

	// Delete Association Record
	// soft delete
	// db.Model(&user).Association("CreditCards").Unscoped().Clear()

	//Permanent Delete
	// db.Unscoped().Model(&user).Association("CreditCards").Unscoped().Clear()
}
