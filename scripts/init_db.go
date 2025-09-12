package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    "orderation/internal/store/mysql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func main() {
    // 加载环境变量
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

    // 创建数据库配置
    config := mysql.NewConfigFromEnv()
    fmt.Printf("连接数据库: %s@%s:%d/%s\n", config.Username, config.Host, config.Port, config.Database)

    // 连接数据库
    db, err := mysql.OpenWithConfig(config)
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer db.Close()

    fmt.Println("数据库连接成功!")

    // 创建数据表
    ctx := context.Background()
    if err := mysql.EnsureSchema(ctx, db); err != nil {
        log.Fatalf("创建数据表失败: %v", err)
    }
    fmt.Println("数据表创建成功!")

    // 插入示例数据
    if err := insertSampleData(ctx, db); err != nil {
        log.Fatalf("插入示例数据失败: %v", err)
    }
    fmt.Println("示例数据插入成功!")

    fmt.Println("数据库初始化完成!")
}

func insertSampleData(ctx context.Context, db *sql.DB) error {
    // 插入示例用户
    users := []struct {
        id, name, email, passHash, role string
    }{
        {"user1", "张三", "zhangsan@example.com", "$2a$10$hash1", "user"},
        {"user2", "李四", "lisi@example.com", "$2a$10$hash2", "user"},
        {"admin1", "管理员", "admin@example.com", "$2a$10$hash3", "admin"},
    }

    for _, user := range users {
        _, err := db.ExecContext(ctx, 
            "INSERT IGNORE INTO users (id, name, email, pass_hash, role) VALUES (?, ?, ?, ?, ?)",
            user.id, user.name, user.email, user.passHash, user.role)
        if err != nil {
            return fmt.Errorf("插入用户 %s 失败: %w", user.name, err)
        }
    }

    // 插入示例餐厅
    restaurants := []struct {
        id, name, address, openTime, closeTime string
    }{
        {"rest1", "川菜馆", "北京市朝阳区xxx街道123号", "09:00", "22:00"},
        {"rest2", "粤菜轩", "上海市浦东新区yyy路456号", "10:00", "21:30"},
        {"rest3", "湘菜坊", "广州市天河区zzz大道789号", "11:00", "23:00"},
    }

    for _, rest := range restaurants {
        _, err := db.ExecContext(ctx,
            "INSERT IGNORE INTO restaurants (id, name, address, open_time, close_time) VALUES (?, ?, ?, ?, ?)",
            rest.id, rest.name, rest.address, rest.openTime, rest.closeTime)
        if err != nil {
            return fmt.Errorf("插入餐厅 %s 失败: %w", rest.name, err)
        }
    }

    // 插入示例桌子
    tables := []struct {
        id, restaurantId, name string
        capacity int
    }{
        {"table1", "rest1", "A01", 4},
        {"table2", "rest1", "A02", 6},
        {"table3", "rest1", "B01", 2},
        {"table4", "rest2", "VIP01", 8},
        {"table5", "rest2", "普通01", 4},
        {"table6", "rest3", "包间01", 10},
        {"table7", "rest3", "大厅01", 6},
    }

    for _, table := range tables {
        _, err := db.ExecContext(ctx,
            "INSERT IGNORE INTO tables (id, restaurant_id, name, capacity) VALUES (?, ?, ?, ?)",
            table.id, table.restaurantId, table.name, table.capacity)
        if err != nil {
            return fmt.Errorf("插入桌子 %s 失败: %w", table.name, err)
        }
    }

    // 插入示例预订
    now := time.Now()
    reservations := []struct {
        id, restaurantId, tableId, userId string
        startTime, endTime time.Time
        guests int
        status string
    }{
        {"resv1", "rest1", "table1", "user1", 
         now.Add(24*time.Hour), now.Add(26*time.Hour), 
         3, "confirmed"},
        {"resv2", "rest2", "table4", "user2", 
         now.Add(48*time.Hour), now.Add(50*time.Hour), 
         6, "pending"},
    }

    for _, resv := range reservations {
        _, err := db.ExecContext(ctx,
            "INSERT IGNORE INTO reservations (id, restaurant_id, table_id, user_id, start_time, end_time, guests, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
            resv.id, resv.restaurantId, resv.tableId, resv.userId, 
            resv.startTime, resv.endTime, resv.guests, resv.status)
        if err != nil {
            return fmt.Errorf("插入预订 %s 失败: %w", resv.id, err)
        }
    }

    return nil
}