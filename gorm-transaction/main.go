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

	// 事务
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&User{Name: "犬夜叉🐶", Age: 100}).Error; err != nil {
			return err
		}
		if err := tx.Create(&User{Name: "戈薇👧", Age: 100}).Error; err != nil {
			return err
		}

		if err := tx.Model(&User{}).Where("name = ?", "犬夜叉🐶").Update("name", "犬夜叉🐺").Error; err != nil {
			return err
		}

		if err := tx.Create(&User{Name: "桔梗👧", Age: 500}).Error; err != nil {
			return err
		}

		// return errors.New("test error")
		return nil
	})

}
