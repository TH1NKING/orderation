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

    // 测试时区修复
    testStart, _ := time.Parse(time.RFC3339, "2025-09-13T02:00:00.000Z")
    testEnd, _ := time.Parse(time.RFC3339, "2025-09-13T04:00:00.000Z")

    fmt.Printf("\nUTC时间: %s - %s\n", testStart.Format("15:04 MST"), testEnd.Format("15:04 MST"))
    
    // 转换为中国时间
    loc, _ := time.LoadLocation("Asia/Shanghai")
    localStart := testStart.In(loc)
    localEnd := testEnd.In(loc)
    
    fmt.Printf("北京时间: %s - %s\n", localStart.Format("15:04 MST"), localEnd.Format("15:04 MST"))

    result := isWithinOperatingHoursFixed(&restaurant, testStart, testEnd)
    fmt.Printf("是否在营业时间内: %t\n", result)

    // 测试营业时间外的情况
    testStart2, _ := time.Parse(time.RFC3339, "2025-09-13T22:00:00.000Z")  // UTC 22:00 = 北京 06:00
    testEnd2, _ := time.Parse(time.RFC3339, "2025-09-14T00:00:00.000Z")   // UTC 00:00 = 北京 08:00
    
    localStart2 := testStart2.In(loc)
    localEnd2 := testEnd2.In(loc)
    
    fmt.Printf("\n测试营业时间外:\n")
    fmt.Printf("UTC时间: %s - %s\n", testStart2.Format("15:04 MST"), testEnd2.Format("15:04 MST"))
    fmt.Printf("北京时间: %s - %s\n", localStart2.Format("15:04 MST"), localEnd2.Format("15:04 MST"))

    result2 := isWithinOperatingHoursFixed(&restaurant, testStart2, testEnd2)
    fmt.Printf("是否在营业时间内: %t (应该是false)\n", result2)
}

// 修复后的函数
func isWithinOperatingHoursFixed(restaurant *models.Restaurant, start, end time.Time) bool {
    fmt.Printf("开始验证营业时间（已修复时区）\n")
    
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
    
    // Convert UTC times to local time (assuming restaurant operates in Asia/Shanghai timezone)
    loc, err := time.LoadLocation("Asia/Shanghai")
    if err != nil {
        // Fallback to UTC+8 if timezone loading fails
        loc = time.FixedZone("CST", 8*3600)
        fmt.Printf("使用固定时区 UTC+8\n")
    } else {
        fmt.Printf("使用Asia/Shanghai时区\n")
    }
    
    localStart := start.In(loc)
    localEnd := end.In(loc)
    
    fmt.Printf("本地时间: %s - %s\n", localStart.Format("15:04"), localEnd.Format("15:04"))
    
    // Get the date and time components in local time
    startDate := localStart.Truncate(24 * time.Hour)
    endDate := localEnd.Truncate(24 * time.Hour)
    
    fmt.Printf("日期范围: %s - %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
    
    // Check each day of the reservation in local time
    for date := startDate; !date.After(endDate); date = date.Add(24 * time.Hour) {
        fmt.Printf("检查日期: %s\n", date.Format("2006-01-02"))
        
        openTime := date.Add(time.Duration(openHour)*time.Hour + time.Duration(openMin)*time.Minute)
        closeTime := date.Add(time.Duration(closeHour)*time.Hour + time.Duration(closeMin)*time.Minute)
        
        fmt.Printf("当日营业时间: %s - %s\n", openTime.Format("15:04"), closeTime.Format("15:04"))
        
        // Handle overnight hours (e.g., 22:00 - 02:00)
        if closeTime.Before(openTime) {
            closeTime = closeTime.Add(24 * time.Hour)
            fmt.Printf("跨夜营业，关闭时间调整为: %s\n", closeTime.Format("15:04"))
        }
        
        // Check if reservation overlaps with this day's operating hours
        dayStart := localStart
        if localStart.Before(date) {
            dayStart = date
        }
        dayEnd := localEnd
        if localEnd.After(date.Add(24*time.Hour)) {
            dayEnd = date.Add(24 * time.Hour)
        }
        
        fmt.Printf("当日预订时间: %s - %s\n", dayStart.Format("15:04"), dayEnd.Format("15:04"))
        
        // If there's any part of the reservation on this day
        if dayStart.Before(dayEnd) {
            fmt.Printf("该日有预订时间段\n")
            // Check if this part is within operating hours
            if dayStart.Before(openTime) || dayEnd.After(closeTime) {
                fmt.Printf("预订时间超出营业时间！开始:%s < 营业开始:%s? %t | 结束:%s > 营业结束:%s? %t\n", 
                    dayStart.Format("15:04"), openTime.Format("15:04"), dayStart.Before(openTime),
                    dayEnd.Format("15:04"), closeTime.Format("15:04"), dayEnd.After(closeTime))
                return false // Any part outside operating hours means rejection
            }
            fmt.Printf("该日预订时间在营业时间内\n")
        } else {
            fmt.Printf("该日无预订时间段\n")
        }
    }
    
    fmt.Printf("所有时间段验证通过\n")
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