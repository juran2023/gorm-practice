package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func prettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic("failed to marshal data: " + err.Error())
	}
	fmt.Println(string(jsonData))
}

// 国家 两个字段
type Country struct {
	ID   uint
	Name string
}

// 地址 (注意：它本身没有自己的数据库表)
type Address struct {
	Street    string
	City      string
	CountryID uint    // 外键，指向国家ID
	Country   Country // Go代码里的关联关系
}

// 公司 id, name, embedded_street, embedded_city_id, embedded_country_id
type Company struct {
	ID      uint
	Name    string
	Address Address `gorm:"embedded"` // 关键点：Address是“嵌入”的
}

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
			Colorful: true,
		},
	)

	dsn := "host=localhost user=postgres password=123456 dbname=embedded_preloading sslmode=disable timezone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Company{}, &Country{})

	// china := Country{
	// 	Name: "China",
	// }

	// usa := Country{
	// 	Name: "USA",
	// }

	// db.Create(&china)
	// db.Create(&usa)

	// company := Company{
	// 	Name: "Apple Corp",
	// 	Address: Address{
	// 		Street:    "100 Home St",
	// 		City:      "New York",
	// 		CountryID: usa.ID,
	// 	},
	// }

	// db.Create(&company)

	// preload
	// 将embedded的结构体的关联关系也加载出来
	var loadedCompany Company
	db.Preload("Address.Country").First(&loadedCompany, 1)
	prettyPrint(loadedCompany)

	var loadedCompany2 Company
	db.First(&loadedCompany2, 1)
	prettyPrint(loadedCompany2)
}
