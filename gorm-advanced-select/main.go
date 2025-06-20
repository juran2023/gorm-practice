package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"default:anonymous"`
	Age      int    `json:"age" gorm:"default:18"`
	LockTest string `json:"lock_test"`
}

func main() {
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
		Logger: newLogger, // 使用配置的日志记录器
	})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移
	db.AutoMigrate(&User{})

	// 确保有测试数据
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		// 创建测试用户
		users := []User{
			{Name: "Test User 1", Age: 25, LockTest: "Initial Value 1"},
			{Name: "Test User 2", Age: 30, LockTest: "Initial Value 2"},
		}
		db.Create(&users)
		fmt.Println("Created test users")
	}

	r := gin.Default()

	// 获取所有用户
	r.GET("/users", func(c *gin.Context) {
		var users []User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})

	// 模拟写锁持有一段时间 - 主要测试接口
	r.PUT("/users/lock/:latency/:id", func(c *gin.Context) {
		id := c.Param("id")
		latencyStr := c.Param("latency")
		latency, _ := strconv.Atoi(latencyStr)
		
		// 记录锁开始时间
		startTime := time.Now()
		
		fmt.Println("尝试获取写锁, ID:", id)
		
		// 简单地开启事务
		tx := db.Begin()
		
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				fmt.Println("事务回滚，原因:", r)
			}
		}()
		
		// 获取写锁 - 关键是使用tx变量和FOR UPDATE
		var user User
		if err := tx.Raw("SELECT * FROM users WHERE id = ? FOR UPDATE", id).Scan(&user).Error; err != nil {
			tx.Rollback()
			fmt.Printf("无法获取锁或记录不存在, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "无法获取记录或加锁",
				"details": err,
			})
			return
		}
		
		fmt.Printf("写锁已获取，将持有 %d 秒, ID: %s\n", latency, id)
		
		// 模拟长时间操作
		time.Sleep(time.Duration(latency) * time.Second)
		
		// 更新数据
		currentTime := time.Now().Format(time.RFC3339)
		if err := tx.Exec("UPDATE users SET lock_test = ? WHERE id = ?", 
			fmt.Sprintf("Updated at %s after %ds lock", currentTime, latency), id).Error; err != nil {
			tx.Rollback()
			fmt.Printf("更新失败, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新失败",
				"details": err,
			})
			return
		}
		
		// 提交事务
		if err := tx.Commit().Error; err != nil {
			fmt.Printf("提交事务失败, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "提交事务失败",
				"details": err,
			})
			return
		}
		
		duration := time.Since(startTime)
		fmt.Printf("锁已释放, ID: %s, 持续时间: %v\n", id, duration)
		
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("成功更新用户 %s", id),
			"lock_held_for": duration.String(),
			"user": user,
		})
	})
	
	// 尝试读取记录 - 演示读锁被写锁阻塞
	r.GET("/users/read/:id", func(c *gin.Context) {
		id := c.Param("id")
		startTime := time.Now()
		
		fmt.Printf("尝试读取记录, ID: %s\n", id)
		
		var user User
		result := db.Raw("SELECT * FROM users WHERE id = ?", id).Scan(&user)
		if result.Error != nil || result.RowsAffected == 0 {
			fmt.Printf("读取失败, ID: %s, 错误: %v\n", id, result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "读取失败",
				"details": result.Error,
				"found": result.RowsAffected > 0,
			})
			return
		}
		
		// 计算读取耗时
		duration := time.Since(startTime)
		fmt.Printf("读取完成, ID: %s, 耗时: %v\n", id, duration)
		
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("读取用户 %s 成功", id),
			"read_time": duration.String(),
			"user": user,
		})
	})
	
	// 尝试快速更新记录 - 演示写锁被写锁阻塞
	r.PUT("/users/quick-update/:id", func(c *gin.Context) {
		id := c.Param("id")
		startTime := time.Now()
		
		fmt.Printf("尝试快速更新, ID: %s\n", id)
		
		// 简单地开启事务
		tx := db.Begin()
		
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				fmt.Println("事务回滚，原因:", r)
			}
		}()
		
		// 尝试获取写锁 - 关键是使用FOR UPDATE
		var user User
		fmt.Printf("尝试获取写锁用于快速更新, ID: %s\n", id)
		if err := tx.Raw("SELECT * FROM users WHERE id = ? FOR UPDATE", id).Scan(&user).Error; err != nil {
			tx.Rollback()
			fmt.Printf("无法获取锁或记录不存在, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "无法获取记录或加锁",
				"details": err,
			})
			return
		}
		
		fmt.Printf("成功获取写锁用于快速更新, ID: %s\n", id)
		
		// 快速更新
		currentTime := time.Now().Format(time.RFC3339)
		if err := tx.Exec("UPDATE users SET lock_test = ? WHERE id = ?", 
			fmt.Sprintf("Quick updated at %s", currentTime), id).Error; err != nil {
			tx.Rollback()
			fmt.Printf("快速更新失败, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "快速更新失败",
				"details": err,
			})
			return
		}
		
		// 提交事务
		if err := tx.Commit().Error; err != nil {
			fmt.Printf("提交事务失败, ID: %s, 错误: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "提交事务失败",
				"details": err,
			})
			return
		}
		
		duration := time.Since(startTime)
		fmt.Printf("快速更新完成, ID: %s, 耗时: %v\n", id, duration)
		
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("快速更新用户 %s 成功", id),
			"operation_time": duration.String(),
			"user": user,
		})
	})

	r.Run(":8080")
}