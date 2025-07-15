package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/sharding"
)

// Order 定义了我们的模型
// 它包含一个分片键 `UserID`
type Order struct {
	ID     int64 `gorm:"primarykey"` // 使用 int64 来接收雪花算法生成的ID
	UserID int64
	Amount uint
}

func main() {

	dns := "host=localhost user=postgres password=123456 dbname=dvdrental port=5432 sslmode=disable timezone=Asia/Shanghai"

	// --- 1. 连接数据库并配置 GORM ---
	// 为了清晰地看到 GORM 生成的 SQL，我们开启了详细日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: 0,
			LogLevel:      logger.Info, // 设置日志级别为 Info，打印所有 SQL
		},
	)
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("无法连接到数据库: " + err.Error())
	}

	// --- 2. 配置并注册分片插件 ---
	// 我们分 4 个片，方便演示
	const NumberOfShards = 4
	shardingMiddleware := sharding.Register(sharding.Config{
		ShardingKey:         "user_id",
		NumberOfShards:      NumberOfShards,
		PrimaryKeyGenerator: sharding.PKSnowflake,
	}, "orders") // 只对 orders 表应用分片规则

	if err := db.Use(shardingMiddleware); err != nil {
		panic("注册分片插件失败: " + err.Error())
	}

	// --- 3. 自动迁移 (创建分片表) ---
	// 插件不负责创建表，我们需要手动为每个分片创建表
	fmt.Println("--- 步骤 3: 准备创建分片表 ---")
	for i := 0; i < NumberOfShards; i++ {
		// 使用 db.Table() 来指定要操作的具体表名
		tableName := fmt.Sprintf("orders_%d", i)
		err := db.Table(tableName).AutoMigrate(&Order{})
		if err != nil {
			panic(fmt.Sprintf("创建分片表 %s 失败: %s", tableName, err.Error()))
		}
	}
	fmt.Println("--- 所有分片表创建成功！---\n")

	// --- 4. 插入数据 ---
	// 我们将创建几个订单，它们的 UserID 会被路由到不同的分片表
	fmt.Println("--- 步骤 4: 插入数据 ---")
	ordersToCreate := []Order{
		{UserID: 1, Amount: 100}, // 1 % 4 = 1  -> orders_1
		{UserID: 2, Amount: 200}, // 2 % 4 = 2  -> orders_2
		{UserID: 4, Amount: 400}, // 4 % 4 = 0  -> orders_0
		{UserID: 5, Amount: 50},  // 5 % 4 = 1  -> orders_1 (和第一个订单在同一个分片)
	}

	for _, order := range ordersToCreate {
		fmt.Printf("--> 正在为 UserID %d 创建订单...\n", order.UserID)
		// 注意这里的 db.Create() 操作的是逻辑上的 "orders" 表
		// 插件会自动将其路由到物理上的 "orders_x" 表
		db.Create(&order)
	}
	fmt.Println("--- 数据插入完成！---\n")

	// --- 5. 查询数据 ---
	fmt.Println("--- 步骤 5: 查询数据 ---")
	var userOrders []Order
	targetUserID := int64(5)

	fmt.Printf("--> 正在查询 UserID 为 %d 的所有订单...\n", targetUserID)
	// 查询时必须带上 ShardingKey (user_id)
	db.Where("user_id = ?", targetUserID).Find(&userOrders)

	fmt.Printf("\n查询结果: 为 UserID %d 找到了 %d 个订单\n", targetUserID, len(userOrders))
	for _, o := range userOrders {
		fmt.Printf("  - 订单 ID: %d, 金额: %d\n", o.ID, o.Amount)
	}
	fmt.Println("--- Demo 结束 ---")
}