package main

import (
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
    db, err := mysql.OpenWithConfig(config)
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    defer db.Close()

    fmt.Println("=== 数据库调试信息 ===")
    
    // 1. 查看所有餐厅
    fmt.Println("\n1. 餐厅列表:")
    rows, _ := db.Query("SELECT id, name FROM restaurants")
    defer rows.Close()
    for rows.Next() {
        var id, name string
        rows.Scan(&id, &name)
        fmt.Printf("  餐厅: %s (%s)\n", name, id)
    }
    
    // 2. 查看rest1的桌子
    fmt.Println("\n2. rest1的桌子:")
    rows2, _ := db.Query("SELECT id, name, capacity FROM tables WHERE restaurant_id = 'rest1'")
    defer rows2.Close()
    for rows2.Next() {
        var id, name string
        var capacity int
        rows2.Scan(&id, &name, &capacity)
        fmt.Printf("  桌子: %s (%s) - 容量: %d\n", name, id, capacity)
    }
    
    // 3. 查看所有预订
    fmt.Println("\n3. 所有预订:")
    rows3, _ := db.Query("SELECT id, restaurant_id, table_id, start_time, end_time, guests, status FROM reservations")
    defer rows3.Close()
    for rows3.Next() {
        var id, rid, tid, status string
        var start, end time.Time
        var guests int
        rows3.Scan(&id, &rid, &tid, &start, &end, &guests, &status)
        fmt.Printf("  预订: %s - 餐厅: %s - 桌子: %s - 时间: %s 到 %s - 人数: %d - 状态: %s\n", 
            id, rid, tid, start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"), guests, status)
    }
    
    // 4. 测试重叠查询
    testStart := time.Date(2025, 9, 13, 18, 0, 0, 0, time.UTC)
    testEnd := time.Date(2025, 9, 13, 20, 0, 0, 0, time.UTC)
    fmt.Printf("\n4. 测试时间段: %s 到 %s\n", testStart.Format("2006-01-02 15:04:05"), testEnd.Format("2006-01-02 15:04:05"))
    
    // 对table1测试重叠查询
    fmt.Println("\n5. table1的重叠预订查询:")
    rows4, _ := db.Query(`
        SELECT id, start_time, end_time, guests, status
        FROM reservations
        WHERE status <> 'cancelled' 
        AND restaurant_id = 'rest1' 
        AND table_id = 'table1'
        AND start_time < ? AND end_time > ?`,
        testEnd, testStart)
    defer rows4.Close()
    
    overlapCount := 0
    for rows4.Next() {
        var id, status string
        var start, end time.Time
        var guests int
        rows4.Scan(&id, &start, &end, &guests, &status)
        fmt.Printf("  重叠预订: %s - 时间: %s 到 %s - 人数: %d - 状态: %s\n", 
            id, start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"), guests, status)
        overlapCount++
    }
    
    if overlapCount == 0 {
        fmt.Println("  ✅ table1 在该时间段无重叠预订，应该可用！")
    } else {
        fmt.Printf("  ❌ table1 在该时间段有 %d 个重叠预订\n", overlapCount)
    }
}