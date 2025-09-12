package main

import (
    "fmt"
    "log"

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

    fmt.Println("=== 数据库中的所有用户 ===")
    
    rows, err := db.Query("SELECT id, name, email, role, created_at FROM users ORDER BY created_at")
    if err != nil {
        log.Fatalf("查询用户失败: %v", err)
    }
    defer rows.Close()

    userCount := 0
    for rows.Next() {
        var id, name, email, role, createdAt string
        err := rows.Scan(&id, &name, &email, &role, &createdAt)
        if err != nil {
            log.Printf("扫描行失败: %v", err)
            continue
        }
        
        userCount++
        fmt.Printf("%d. 用户: %s (%s)\n", userCount, name, email)
        fmt.Printf("   ID: %s\n", id)
        fmt.Printf("   角色: %s\n", role)
        fmt.Printf("   创建时间: %s\n", createdAt)
        fmt.Println()
    }
    
    if userCount == 0 {
        fmt.Println("❌ 数据库中没有任何用户！")
    } else {
        fmt.Printf("✅ 总共找到 %d 个用户\n", userCount)
    }
    
    // 检查是否有管理员
    adminRows, err := db.Query("SELECT name, email FROM users WHERE role = 'admin'")
    if err != nil {
        log.Printf("查询管理员失败: %v", err)
        return
    }
    defer adminRows.Close()
    
    fmt.Println("\n=== 管理员用户 ===")
    adminCount := 0
    for adminRows.Next() {
        var name, email string
        adminRows.Scan(&name, &email)
        adminCount++
        fmt.Printf("管理员 %d: %s (%s)\n", adminCount, name, email)
    }
    
    if adminCount == 0 {
        fmt.Println("❌ 数据库中没有管理员用户！")
        fmt.Println("\n💡 建议：")
        fmt.Println("1. 检查环境变量 ADMIN_EMAIL 和 ADMIN_PASSWORD 是否设置")
        fmt.Println("2. 确保服务器启动时会调用 BootstrapAdmin 函数")
    } else {
        fmt.Printf("✅ 找到 %d 个管理员\n", adminCount)
    }
}