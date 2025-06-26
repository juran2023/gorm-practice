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

	// 删除一条记录
	user := User{Model: gorm.Model{ID: 623}}
	db.Delete(&user)

	// 根据主键删除
	db.Delete(&User{}, 624)

	db.Delete(&User{}, []int{625, 626})

	// 钩子函数 - 正确捕获和检查错误
	adminUser := User{Model: gorm.Model{ID: 627}, Role: "admin"}
	result := db.Delete(&adminUser) // 将结果保存到 result 变量中

	// 检查 result.Error 是否有值
	if result.Error != nil {
		fmt.Printf("成功捕获到错误: %v\n", result.Error)
		fmt.Printf("受影响的行数: %d\n", result.RowsAffected)
	} else {
		fmt.Println("错误：没有按预期捕获到来自Hook的错误。")
	}

	// 批量删除
	// db.Where("name LIKE ?", "马%").Delete(&User{})
	db.Delete(&User{})

	// db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})

	// 返回删除行的数据
	user = User{Model: gorm.Model{ID: 626}}
	db.Clauses(clause.Returning{}).Delete(&user)

	var byteArr []byte
	var err2 error

	// 打印user的json
	byteArr, err2 = json.Marshal(user)
	if err2 != nil {
		fmt.Println("json.Marshal error:", err2)
	}
	fmt.Println("user json:", string(byteArr))

	// 如果你并不想嵌套gorm.Model，你也可以像下方例子那样开启软删除特性：
	type Actor struct {
		ID      int
		Deleted gorm.DeletedAt
		Name    string
	}

	// 查找被软删除的记录
	// 获取查出的记录
	user = User{}
	db.Unscoped().Where("id = ?", 611).Find(&user)
	byteArr, err2 = json.Marshal(user)
	if err2 != nil {
		fmt.Println("json.Marshal error:", err2)
	}
	fmt.Println("user json:", string(byteArr))

	// 彻底删除一条记录
	db.Unscoped().Delete(&User{}, 628)

	// 提示 当使用DeletedAt创建唯一复合索引时，你必须使用其他的数据类型，例如通过gorm.io/plugin/soft_delete插件将字段类型定义为unix时间戳等等

	// import "gorm.io/plugin/soft_delete"

	//	type User struct {
	//	  ID        uint
	//	  Name      string                `gorm:"uniqueIndex:udx_name"`
	//	  DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:udx_name"`
	//	}

}
