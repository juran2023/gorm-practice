package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"log"
	"os"
	"time"
	"gorm.io/gorm/logger"
	"net/http"
	"fmt"
	"database/sql/driver"
	"encoding/json"
)

type User struct {
  gorm.Model
  Name string `json:"name" binding:"required" gorm:"default:anonymous"`
  Age int `json:"age" binding:"required,gt=0" gorm:"default:18"`
  Birthday Date `gorm:"type:date" json:"birthday" binding:"required"`
}

type Date time.Time

func (d *Date) UnmarshalJSON(b []byte) error {
  var dateStr string
  if err := json.Unmarshal(b, &dateStr); err != nil {
    return err
  }

  date, err := time.Parse("2006-01-02", dateStr)
  if err != nil {
    return err
  }
  *d = Date(date)
  return nil
}

func (d Date) MarshalJSON() ([]byte, error) {

  return json.Marshal(time.Time(d).Format("2006-01-02"))
}

func (d Date) Value() (driver.Value, error){
  return time.Time(d), nil
}

func (d *Date) Scan(value interface{}) error {
  v, ok := value.(time.Time)
  if !ok {
    return fmt.Errorf("invalid type %T for Date", value)
  }
  *d = Date(v)
  return nil
}



func main(){
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel: logger.Info,
			Colorful: true,
		},
	)
	gin.ForceConsoleColor()

	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=123456 dbname=dvdrental port=5432 sslmode=disable timezone=Asia/Shanghai"), &gorm.Config{
		Logger: newLogger,
		CreateBatchSize: 1000,
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})

	r := gin.Default()

	// 查询所有用户
	r.GET("/users", func(c *gin.Context){
		var users []User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})

	// 根据名字查询一个用户
	r.GET("/users/:name", func(c *gin.Context){
		name := c.Param("name")
		var user User
		if err := db.Where("name = ?", name).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	})
	r.Run(":8080")
}