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
	"gorm.io/hints"
)

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"default:anonymous"`
	Age      int    `json:"age" gorm:"default:18"`
	LockTest string `json:"lock_test"`
	Role     string `json:"role" gorm:"default:user"`
}

func AgeGreaterThan(age int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("age > ?", age)
	}
}

func NameLengthGreaterThan(length int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("LENGTH(name) > ?", length)
	}
}

func NamesIn(names []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name IN ?", names)
	}
}

// 查询钩子
func (u *User) AfterFind(tx *gorm.DB) (err error) {
	if u.Role == "user" {
		u.Role = "admin"
	}
	return
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

	// 为属性使用 Assign with result
	user = &User{}

	db.Where(User{Name: "Pain"}).Assign(User{Age: 100}).FirstOrInit(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 为属性使用 Assign without result
	user = &User{}

	db.Where(User{Name: "anonymous"}).Assign(User{Age: 100}).FirstOrInit(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// FirstOrCreate
	// FirstOrCreate 用于获取与特定条件匹配的第一条记录，或者如果没有找到匹配的记录，创建一个新的记录。 这个方法在结构和map条件下都是有效的。
	db.Where(User{Name: "anonymous", Age: 100}).FirstOrCreate(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 配合 Attrs 使用 FirstOrCreate with result
	// 找到结果，忽略Attrs
	user = &User{}
	db.Where(User{Name: "anonymous"}).Attrs(User{Age: 1000}).FirstOrCreate(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 配合 Attrs 使用 FirstOrCreate without result
	user = &User{}
	db.Where(User{Name: "kiwi"}).Attrs(User{Age: 999, LockTest: "test"}).FirstOrCreate(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 配合 Assign 使用 FirstOrCreate 保存
	user = &User{}
	db.Where(User{Name: "kikawa"}).Assign(User{Age: 999, LockTest: "test"}).FirstOrCreate(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 配合 Assign 使用 FirstOrCreate 更新
	user = &User{}
	db.Where(User{Name: "仙道"}).Assign(User{Age: 16, LockTest: "test"}).FirstOrCreate(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 优化器、索引提示
	user = &User{}
	db.Clauses(hints.New("MAX_EXECUTION_TIME(10000)")).Where("name = ?", "kikawa").FirstOrInit(&user)
	jsonBytes, err2 = json.MarshalIndent(user, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// 索引提示
	// 对指定索引提供建议
	// db.Clauses(hints.UseIndex("idx_user_name")).Find(&User{})
	// SQL: SELECT * FROM `users` USE INDEX (`idx_user_name`)

	// 强制对JOIN操作使用某些索引
	// db.Clauses(hints.ForceIndex("idx_user_name", "idx_user_id").ForJoin()).Find(&User{})
	// SQL: SELECT * FROM `users` FORCE INDEX FOR JOIN (`idx_user_name`,`idx_user_id`)

	// 迭代
	rs, err := db.Model(&User{}).Rows()
	if err != nil {
		log.Fatalf("迭代失败: %v", err)
	}
	defer rs.Close()

	for rs.Next() {
		user = &User{}
		db.ScanRows(rs, user)
		fmt.Println(user)
	}

	// FindInBatches
	// 处理记录，批处理大小为100
	// result := db.Where("processed = ?", false).FindInBatches(&results, 100, func(tx *gorm.DB, batch int) error {
	//   for _, result := range results {
	//     // 对批中的每条记录进行操作
	//   }

	// 保存对当前批记录的修改
	//   tx.Save(&results)

	// tx.RowsAffected 提供当前批处理中记录的计数（the count of records in the current batch）
	// 'batch' 变量表示当前批号（the current batch number）

	// 返回 error 将阻止更多的批处理
	//   return nil
	// })

	// result.Error 包含批处理过程中遇到的任何错误
	// result.RowsAffected 提供跨批处理的所有记录的计数（the count of all processed records across batches）

	// Pluck
	var names []string
	db.Model(&User{}).Pluck("name", &names)
	fmt.Println(names)

	var ages []int
	db.Model(&User{}).Pluck("age", &ages)
	fmt.Println(ages)

	db.Model(&User{}).Distinct().Pluck("name", &names)
	fmt.Println(names)

	db.Model(&User{}).Distinct().Pluck("age", &ages)
	fmt.Println(ages)

	// Scope
	db.Scopes(AgeGreaterThan(30), NameLengthGreaterThan(5)).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	db.Scopes(NamesIn([]string{"Pain", "knight"})).Find(&users)
	jsonBytes, err2 = json.MarshalIndent(users, "", "  ")
	if err2 != nil {
		log.Fatalf("序列化失败: %v", err2)
	}
	fmt.Println(string(jsonBytes))

	// Count
	var count int64
	db.Model(&User{}).Count(&count)
	fmt.Println("count: ", count)

	db.Model(&User{}).Group("CHAR_LENGTH(name)").Count(&count)
	fmt.Println("char_length count: ", count)

	db.Table("users").Select("COUNT(DISTINCT CHAR_LENGTH(name))").Count(&count)
	fmt.Println("distinct char_length count: ", count)

}
