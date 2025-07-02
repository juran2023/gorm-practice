package main

import (
	"errors"
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
		Logger:         newLogger,
		TranslateError: true,
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{}, &Language{}, &UserLanguage{})

	var user User
	err = db.First(&user, 99).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		//	panic("record not found!")
	}

	// handle error code
	// user = User{
	// 	Model: gorm.Model{ID: 1},
	// 	Name:  "MakaBaka",
	// 	Age:   100,
	// }

	// err = db.Create(&user).Error

	// if err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) {
	// 		switch pgErr.Code {
	// 		case "23505":
	// 			panic("duplicate key error")
	// 		case "23503":
	// 			panic("foreign key constraint violation")
	// 		default:
	// 			panic("unknown error")
	// 		}
	// 	}
	// }

	// 方言转换错误
	user = User{
		Model: gorm.Model{ID: 1},
		Name:  "MakaBaka",
		Age:   100,
	}

	err = db.Create(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			panic("duplicate key error")
		}
	}

}
