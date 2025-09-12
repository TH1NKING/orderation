package main

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"

    "orderation/internal/store/mysql"
    "orderation/internal/models"
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

    // 获取餐厅信息
    row := db.QueryRow("SELECT id, name, open_time, close_time FROM restaurants WHERE id = ?", "rest1")
    var restaurant models.Restaurant
    err = row.Scan(&restaurant.ID, &restaurant.Name, &restaurant.OpenTime, &restaurant.CloseTime)
    if err != nil {
        log.Fatalf("获取餐厅信息失败: %v", err)
    }

    fmt.Printf("餐厅: %s\n", restaurant.Name)
    fmt.Printf("营业时间: %s - %s\n", restaurant.OpenTime, restaurant.CloseTime)

    // 测试时间范围
    testStart, _ := time.Parse(time.RFC3339, "2025-09-13T06:00:00.000Z")
    testEnd, _ := time.Parse(time.RFC3339, "2025-09-13T08:00:00.000Z")

    fmt.Printf("\n测试预订时间: %s - %s\n", testStart.Format("15:04"), testEnd.Format("15:04"))

    result := isWithinOperatingHours(&restaurant, testStart, testEnd)
    fmt.Printf("是否在营业时间内: %t\n", result)

    // 测试深夜时间
    testStart2, _ := time.Parse(time.RFC3339, "2025-09-13T23:00:00.000Z")
    testEnd2, _ := time.Parse(time.RFC3339, "2025-09-14T01:00:00.000Z")

    fmt.Printf("\n测试预订时间2: %s - %s\n", testStart2.Format("15:04"), testEnd2.Format("15:04"))

    result2 := isWithinOperatingHours(&restaurant, testStart2, testEnd2)
    fmt.Printf("是否在营业时间内: %t\n", result2)
}

// 从handlers包复制的函数
func isWithinOperatingHours(restaurant *models.Restaurant, start, end time.Time) bool {
    // Parse operating hours (format: "09:00")
    openHour, openMin, err := parseTime(restaurant.OpenTime)
    if err != nil {
        fmt.Printf("解析开始时间失败: %v\n", err)
        return false
    }
    closeHour, closeMin, err := parseTime(restaurant.CloseTime)
    if err != nil {
        fmt.Printf("解析结束时间失败: %v\n", err)
        return false
    }
    
    fmt.Printf("营业时间解析: 开始 %d:%d, 结束 %d:%d\n", openHour, openMin, closeHour, closeMin)
    
    // Get the date and time components
    startDate := start.Truncate(24 * time.Hour)
    endDate := end.Truncate(24 * time.Hour)
    
    fmt.Printf("预订日期范围: %s - %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
    
    // Check each day of the reservation
    for date := startDate; !date.After(endDate); date = date.Add(24 * time.Hour) {
        openTime := date.Add(time.Duration(openHour)*time.Hour + time.Duration(openMin)*time.Minute)
        closeTime := date.Add(time.Duration(closeHour)*time.Hour + time.Duration(closeMin)*time.Minute)
        
        fmt.Printf("日期 %s: 营业 %s - %s\n", date.Format("2006-01-02"), openTime.Format("15:04"), closeTime.Format("15:04"))
        
        // Handle overnight hours (e.g., 22:00 - 02:00)
        if closeTime.Before(openTime) {
            closeTime = closeTime.Add(24 * time.Hour)
            fmt.Printf("跨夜营业，关闭时间调整为: %s\n", closeTime.Format("15:04"))
        }
        
        // Check if reservation overlaps with this day's operating hours
        dayStart := start
        if start.Before(date) {
            dayStart = date
        }
        dayEnd := end
        if end.After(date.Add(24*time.Hour)) {
            dayEnd = date.Add(24 * time.Hour)
        }
        
        fmt.Printf("当日预订时间: %s - %s\n", dayStart.Format("15:04"), dayEnd.Format("15:04"))
        
        if dayStart.Before(closeTime) && dayEnd.After(openTime) {
            fmt.Printf("与营业时间有重叠\n")
            // There's an overlap, check if it's within operating hours
            if dayStart.Before(openTime) || dayEnd.After(closeTime) {
                fmt.Printf("预订时间超出营业时间！\n")
                return false
            }
            fmt.Printf("预订时间在营业时间内\n")
        } else {
            fmt.Printf("与营业时间无重叠\n")
        }
    }
    
    return true
}

func parseTime(timeStr string) (hour, minute int, err error) {
    parts := strings.Split(timeStr, ":")
    if len(parts) != 2 {
        return 0, 0, fmt.Errorf("invalid time format")
    }
    
    hour, err = strconv.Atoi(parts[0])
    if err != nil {
        return 0, 0, err
    }
    
    minute, err = strconv.Atoi(parts[1])
    if err != nil {
        return 0, 0, err
    }
    
    return hour, minute, nil
}