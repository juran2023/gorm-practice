package main

import (
  "fmt"
  "database/sql/driver"
  "encoding/json"
  "gorm.io/gorm"
  "gorm.io/driver/postgres"
  "github.com/gin-gonic/gin"
  "net/http"
  "time"
  "gorm.io/gorm/logger"
  "log"
  "os"
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
  // 配置 GORM 日志
  newLogger := logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
    logger.Config{
      SlowThreshold: time.Second,   // 慢 SQL 阈值
      LogLevel:      logger.Info,   // 日志级别
      Colorful:      true,          // 彩色打印
    },
  )

  dns := "host=localhost user=postgres password=123456 dbname=dvdrental port=5432 sslmode=disable timezone=Asia/Shanghai"
  db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
    CreateBatchSize: 1000,
    Logger: newLogger, // 使用配置的日志记录器
  })
  if err != nil {
    panic("failed to connect database")
  }

  db.AutoMigrate(&User{})

  r := gin.Default()
  r.POST("/users", func(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    if err := db.Create(&user).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusCreated, user)
  })
  
  // 批量新增用户
  r.POST("/users/batch", func(c *gin.Context){
    var users []*User
    if err := c.ShouldBindJSON(&users); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if err := db.Create(&users).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusCreated, users)
  })

  // 指定字段新增
  r.POST("/users/partial", func(c *gin.Context){
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if err := db.Select("Name", "Birthday").Create(&user).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusCreated, user)
  })

  // 忽略指定字段新增
  r.POST("/users/ignore", func(c *gin.Context){
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if err := db.Omit("Age").Create(&user).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusCreated, user)
  })

  // 测试CreateInBatches
  r.POST("/users/batch/in-batches", func(c *gin.Context){
    // 自己模拟200条数据
    var users []*User
    for i := 0; i < 200; i++ {
      users = append(users, &User{
        Name: fmt.Sprintf("User %d", i),
        Age: i + 1,
        Birthday: Date(time.Now()),
      })
    }

    if err := db.CreateInBatches(users, 100).Error; err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusCreated, gin.H{"message": "200 users created"})
  })


  r.Run(":8080")
}