package main

import (
	"encoding/json"
	"errors"
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

func (u *User) BeforeDelete(tx *gorm.DB) (err error) {
	fmt.Println("==============================触发了BeforeDelete==============================")
	if u.Role == "admin" {
		return errors.New("admin cannot be deleted")
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

	// Row & Rows
	var name string
	var age int
	row := db.Table("users").Where("name = ?", "马飞飞").Select("name", "age").Row()
	row.Scan(&name, &age)
	fmt.Println("name:", name, "age:", age)

	rows, err := db.Table("users").Where("name = ?", "马飞飞").Select("name", "age").Rows()
	if err != nil {
		fmt.Println("Rows error:", err)
	}
	defer rows.Close()

	fmt.Println("==============================Rows==============================")
	var byteArr []byte

	for rows.Next() {
		var user User
		rows.Scan(&name, &age)
		// 将 sql.Rows 扫描至 model
		db.ScanRows(rows, &user)
		fmt.Println("name:", name, "age:", age)
		byteArr, err = json.Marshal(user)
		if err != nil {
			fmt.Println("json.Marshal error:", err)
		}
		fmt.Println("user json:", string(byteArr))
	}

	// 在一条 tcp DB 连接中运行多条 SQL (不是事务)
	db.Connection(func(tx *gorm.DB) error {
		tx.Exec("SELECT 1")
		tx.Exec("SELECT 2")
		return nil
	})

}
