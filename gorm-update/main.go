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
	Name     string    `json:"name" gorm:"default:anonymous"`
	Age      int       `json:"age" gorm:"default:18"`
	Birthday time.Time `json:"birthday"`
	LockTest string    `json:"lock_test"`
	Role     string    `json:"role" gorm:"default:user"`
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

	// 保存所有字段
	// user := User{
	// 	Name:     "Adam",
	// 	Age:      20,
	// 	LockTest: "test",
	// }
	// db.Save(&user)

	// db.Save(&User{Model: gorm.Model{ID: user.ID}})

	// 更新单个列
	db.Model(&User{}).Where("name = ?", "仙道").Update("lock_test", time.Now().Format(time.RFC3339))

	user := User{
		Model: gorm.Model{ID: 611},
		Name:  "kikawa",
	}

	db.Model(&user).Where("name = ?", "kiko").Update("lock_test", time.Now().Format(time.RFC3339))

	// 更新多列 指定更新列
	db.Model(&user).Select("birthday").Updates(map[string]interface{}{
		"age":      100,
		"birthday": time.Now(),
	})

	// 更新多列 忽略更新特定列
	db.Model(&user).Omit("birthday").Updates(map[string]interface{}{
		"age":      100,
		"birthday": time.Now(),
	})

	// 更新多列
	db.Model(&user).Select("birthday").Updates(map[string]interface{}{
		"age":      100,
		"birthday": time.Now(),
	})

}
