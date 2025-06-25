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
	"gorm.io/gorm/clause"
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

// 更新Hook
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if u.Role == "user" {
		return errors.New("users are not allowed to update")
	}
	return
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
	db.Model(&user).Select("birthday", "age").Updates(map[string]interface{}{
		"age":      102,
		"birthday": time.Now().AddDate(0, 0, 1),
	})

	// 更写多列 但不设置值
	db.Model(&user).Select("birthday", "age").Updates(map[string]interface{}{
		"age":      103,
		"birthday": nil,
	})

	// 更新多列 不设置birthday 会发生什么？
	db.Model(&user).Select("birthday", "age").Updates(map[string]interface{}{
		"age": gorm.Expr("age + ?", 1),
	})

	// 选择所有字段
	db.Model(&user).Select("*").Updates(map[string]interface{}{
		"age": gorm.Expr("age + ?", 1),
	})

	// 选择所有字段 除了birthday
	db.Model(&user).Select("*").Omit("age").Updates(map[string]interface{}{
		"age": gorm.Expr("age + ?", 1),
	})

	// --- 演示如何触发更新 Hook ---
	fmt.Println("\n--- 演示触发Hook ---")
	// 1. 首先，根据ID从数据库中查找一个完整的用户实例
	var userToUpdate User
	// 我们假设更新 ID 为 611 的用户，并且他的 role 是 'user'
	if err := db.First(&userToUpdate, 611).Error; err != nil {
		fmt.Println("找不到用于演示Hook的用户:", err)
	} else {
		// 2. 尝试更新这个实例的 Age 字段
		// 因为我们是在一个加载的实例上调用 Update，所以会触发 Hook
		fmt.Printf("尝试更新用户 %d (Role: %s)，这应该会触发BeforeUpdate Hook...\n", userToUpdate.ID, userToUpdate.Role)
		result := db.Model(&userToUpdate).Update("age", 999)

		// 3. 检查是否收到了来自 Hook 的错误
		if result.Error != nil {
			fmt.Printf("成功触发Hook并捕获到错误: %v\n", result.Error)
		} else {
			fmt.Println("Hook没有按预期返回错误。")
		}
	}

	// 批量更新
	db.Model(&User{}).Where("role = ?", "user").Updates(User{
		Age: 100,
	})

	// Update with map
	db.Model(&User{}).Where("id IN ?", []int{611}).Updates(map[string]interface{}{
		"age":      101,
		"birthday": time.Now().AddDate(1, 0, 0),
	})

	// 尝试全局更新
	db.Model(&User{}).Update("age", 102)

	db.Exec("UPDATE users SET age = ?", 103)

	//允许全局更新
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&User{}).Update("age", 104)

	// 更新的记录数
	results := db.Model(&User{}).Where("name = ?", "kiwi").Update("age", 105)
	fmt.Println(results.RowsAffected)

	// 高级选项
	//使用 SQL 表达式更新
	db.Model(&user).Update("age", gorm.Expr("age * ? + ?", 2, 100))

	// 更新多列
	db.Model(&user).Updates(map[string]interface{}{
		"age":      gorm.Expr("age * ? + ?", 2, 100),
		"birthday": gorm.Expr("birthday + INTERVAL '1 day'"),
	})

	// UpdateColumn + SQL 表达式更新
	db.Model(&user).UpdateColumn("age", gorm.Expr("age * ? + ?", 2, 100))

	// 根据子查询进行更新
	db.Model(&User{}).Where("role = ?", "user").Update("age", db.Model(&user).Select("age"))

	// db.Table("users as u").Where("name = ?", "jinzhu").Update("company_name", db.Table("companies as c").Select("name").Where("c.id = u.company_id"))

	// db.Table("users as u").Where("name = ?", "jinzhu").Updates(map[string]interface{}{"company_name": db.Table("companies as c").Select("name").Where("c.id = u.company_id")})

	//不使用 Hook 和时间追踪
	// 如果你希望更新时跳过 Hook 方法，并且不追踪更新的时间，你可以使用 UpdateColumn, UpdateColumns
	// 打印user
	var jsonBytes []byte
	var err2 error
	// 打印更改前的user
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))
	db.Model(&user).UpdateColumn("age", 107)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 返回修改行的数据
	db.Model(&user).Clauses(clause.Returning{}).Update("age", 108)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	db.Model(&user).Clauses(clause.Returning{
		Columns: []clause.Column{
			{Name: "age"},
			{Name: "birthday"},
			{Name: "name"},
		},
	}).Update("age", 109)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// returning的力量
	var emptyUser User
	db.Model(&emptyUser).Clauses(clause.Returning{}).Where("id = ?", 3).Update("age", 110)
	jsonBytes, err2 = json.MarshalIndent(emptyUser, "", "  ")
	if err2 != nil {
		log.Fatalf("json.Marshal failed: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	//

}
