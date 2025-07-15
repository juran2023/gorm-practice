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

	// 创建一个viewd
	// query := db.Model(&User{}).Where("age > ?", 20)

	// Using GORM's CreateView method (not working correctly)
	// db.Migrator().CreateView("users_view", gorm.ViewOption{Query: query, CheckOption: "With Check Option"})

	// Instead, use raw SQL to create the view with CHECK OPTION
	db.Exec("DROP VIEW IF EXISTS users_view")
	db.Exec(`CREATE VIEW users_view AS 
         SELECT * FROM users 
         WHERE age > 20
         WITH CHECK OPTION`)

	// 通过视图插入

	// Test inserting through view
	validUser := User{Name: "Valid User", Age: 25}
	err = db.Table("users_view").Create(&validUser).Error
	if err != nil {
		log.Printf("Failed to insert valid user: %v", err)
	} else {
		log.Println("Successfully inserted valid user")
	}

	invalidUser := User{Name: "Invalid User", Age: 15}
	err = db.Table("users_view").Create(&invalidUser).Error
	if err != nil {
		log.Printf("Failed to insert invalid user (expected): %v", err)
	} else {
		log.Println("Unexpectedly inserted invalid user")
	}

	// First, insert a user directly to the table for update testing
	testUser := User{Name: "Test User", Age: 30}
	db.Create(&testUser)

	// Test updating through view to valid age
	err = db.Table("users_view").Model(&User{}).Where("id = ?", testUser.ID).Update("age", 35).Error
	if err != nil {
		log.Printf("Failed to update to valid age: %v", err)
	} else {
		log.Println("Successfully updated to valid age")
	}

	// Test updating through view to invalid age
	err = db.Table("users_view").Model(&User{}).Where("id = ?", testUser.ID).Update("age", 10).Error
	if err != nil {
		log.Printf("Failed to update to invalid age (expected): %v", err)
	} else {
		log.Println("Unexpectedly updated to invalid age")
	}

	// Let's also check if the view definition is correct
	var viewDef string
	db.Raw("SELECT definition FROM pg_views WHERE viewname = 'users_view'").Scan(&viewDef)
	log.Printf("View definition: %s", viewDef)
}
