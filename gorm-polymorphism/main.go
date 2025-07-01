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

type Dog struct {
	gorm.Model
	Name  string
	Foods []Food `gorm:"polymorphic:Species;"`
}

type Cat struct {
	gorm.Model
	Name  string
	Foods []Food `gorm:"polymorphic:Species;"`
}

type Human struct {
	gorm.Model
	Name  string
	Foods []Food `gorm:"polymorphic:Species;"`
}

type Food struct {
	gorm.Model
	Name        string
	SpeciesID   int
	SpeciesType string
}

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	// 写一个polymorphism的例子
	dsn := "host=localhost user=postgres password=123456 timezone=Asia/Shanghai dbname=dvdrental"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Dog{}, &Cat{}, &Human{}, &Food{})

	// dog := Dog{
	// 	Name: "dog",
	// 	Foods: []Food{
	// 		{Name: "bone"},
	// 		{Name: "milk"},
	// 	},
	// }

	// db.Create(&dog)

	// cat := Cat{
	// 	Name: "cat",
	// 	Foods: []Food{
	// 		{Name: "fish"},
	// 		{Name: "milk"},
	// 	},
	// }
	// db.Create(&cat)

	// human := Human{
	// 	Name: "human",
	// 	Foods: []Food{
	// 		{Name: "apple"},
	// 		{Name: "banana"},
	// 		{Name: "orange"},
	// 	},
	// }
	// db.Create(&human)

	// 查询
	var human Human
	db.Model(&Human{}).Preload("Foods").Where("id = ?", 2).Find(&human)

	var jsonData []byte
	jsonData, err = json.MarshalIndent(human, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}
