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

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"default:anonymous"`
	Age      int    `json:"age" gorm:"default:18"`
	LockTest string `json:"lock_test"`
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

	dns := "host=localhost user=postgres password=123456 dbname=dvdrental port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})

	var jsonBytes []byte
	var err2 error

	// 测试子查询
	var users []*User
	subQuery := db.Model(&User{}).Select("AVG(age)")
	db.Select("name", "age").Where("age > (?)", subQuery).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 测试From 单个子查询
	users = []*User{} // 清空切片，避免之前查询的结果影响
	rows, err := db.Table("(?) as u", db.Model(&User{}).Select("name", "age")).Rows()
	if err != nil {
		log.Fatalf("db.Table failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		// 只扫描name和age字段
		if err := rows.Scan(&user.Name, &user.Age); err != nil {
			log.Printf("rows.Scan failed: %v", err)
			continue
		}
		users = append(users, &user)
	}

	fmt.Printf("查询到 %d 条记录\n", len(users))
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 测试From 多个子查询
	users = []*User{} // 清空切片，避免之前查询的结果影响
	subQuery1 := db.Model(&User{}).Select("name")
	subQuery2 := db.Model(&User{}).Select("name")

	// 定义一个临时结构体来接收查询结果
	type Result struct {
		Name1 string
		Name2 string
	}

	var results []Result
	db.Table("(?) as u1, (?) as u2", subQuery1, subQuery2).Select("u1.name as Name1, u2.name as Name2").Scan(&results)

	// 打印map内容
	jsonBytes, err2 = json.MarshalIndent(results, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	db.Where(
		db.Where("age <= ?", 18).Where("name = ?", "Pain"),
	).Or(
		db.Where("age > ?", 18).Where(
			db.Where("name = ?", "knight").Or("name = ?", "joe biden"),
		),
	).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 带多个列的 In
	db.Where("(name, age) IN (?)", [][]interface{}{
		{"Pain", 18},
		{"knight", 23},
	}).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 命名参数 - 不使用
	db.Where("name = ? or name = ?", "Pain", "knight").Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 命名参数 - 使用
	db.Where("name = @name or name = @name2", map[string]interface{}{
		"name":  "仙道",
		"name2": "Tianma😭",
	}).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// Find 至 map
	var result []map[string]interface{}
	db.Model(&User{}).Find(&result)
	jsonBytes, err2 = json.MarshalIndent(result, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// FirstOrInit
	user := &User{
		Name: "Rebecca",
	}

	db.FirstOrInit(&user, map[string]interface{}{
		"name": "Rebecca",
		"age":  17,
	})
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 初始化user
	user = &User{}
	// 使用 Attrs 进行初始化
	db.Where(User{Name: "Rebecca"}).Attrs(User{Age: 17}).FirstOrInit(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

}
