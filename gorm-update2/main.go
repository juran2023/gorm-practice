package main

import (
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
	Name     string    `json:"name" gorm:"default:anonymous"`
	Age      int       `json:"age" gorm:"default:18"`
	Birthday time.Time `json:"birthday"`
	LockTest string    `json:"lock_test"`
	Role     string    `json:"role" gorm:"default:user"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	fmt.Println("==============================触发了BeforeSave==============================")
	// 检查创建操作的源是 map 还是 struct
	if destMap, isMap := tx.Statement.Dest.(map[string]interface{}); isMap {
		// 如果是 map，我们直接修改 map 中的值，而不是使用 SetColumn
		if age, ok := destMap["age"].(int); ok {
			destMap["age"] = age + 20
		}
	} else {
		// 如果是 struct，直接修改实例的字段值
		u.Age += 20
	}
	return nil
}

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logger.Info,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
		},
	)

	dsn := "host=localhost user=postgres password=123456 dbname=dvdrental port=5432 sslmode=disable timezone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Create
	newUser := User{
		Name: "马飞飞",
		Age:  100,
	}
	db.Create(&newUser)
	fmt.Println(newUser.ID)

	// Create with map
	db.Model(&User{}).Create(map[string]interface{}{
		"name": "马飞飞",
		"age":  100,
	})

}
