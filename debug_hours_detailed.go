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

    // 测试多个时间场景
    testCases := []struct {
        name      string
        startTime string
        endTime   string
        expected  bool
    }{
        {"早上6-8点", "2025-09-13T06:00:00.000Z", "2025-09-13T08:00:00.000Z", false},
        {"晚上23-01点", "2025-09-13T23:00:00.000Z", "2025-09-14T01:00:00.000Z", false},
        {"上午10-12点", "2025-09-13T10:00:00.000Z", "2025-09-13T12:00:00.000Z", true},
        {"下午18-20点", "2025-09-13T18:00:00.000Z", "2025-09-13T20:00:00.000Z", true},
        {"跨营业时间8-10点", "2025-09-13T08:00:00.000Z", "2025-09-13T10:00:00.000Z", false},
        {"跨营业时间21-23点", "2025-09-13T21:00:00.000Z", "2025-09-13T23:00:00.000Z", false},
    }

    for _, tc := range testCases {
        fmt.Printf("\n=== 测试: %s ===\n", tc.name)
        
        start, _ := time.Parse(time.RFC3339, tc.startTime)
        end, _ := time.Parse(time.RFC3339, tc.endTime)
        
        fmt.Printf("UTC时间: %s - %s\n", start.Format("15:04 MST"), end.Format("15:04 MST"))
        
        // 转换为中国时间显示
        chinaLoc, _ := time.LoadLocation("Asia/Shanghai")
        startChina := start.In(chinaLoc)
        endChina := end.In(chinaLoc)
        fmt.Printf("中国时间: %s - %s\n", startChina.Format("15:04 MST"), endChina.Format("15:04 MST"))

        result := isWithinOperatingHours(&restaurant, start, end)
        status := "✅ 通过"
        if !result {
            status = "❌ 拒绝"
        }
        
        expected := "✅ 应通过"
        if !tc.expected {
            expected = "❌ 应拒绝"
        }
        
        fmt.Printf("结果: %s | 预期: %s", status, expected)
        if result == tc.expected {
            fmt.Printf(" | 🎯 正确")
        } else {
            fmt.Printf(" | 🐛 错误")
        }
        fmt.Println()
    }
}

func isWithinOperatingHours(restaurant *models.Restaurant, start, end time.Time) bool {
    fmt.Printf("  调试: 开始验证营业时间\n")
    
    // Parse operating hours (format: "09:00")
    openHour, openMin, err := parseTime(restaurant.OpenTime)
    if err != nil {
        fmt.Printf("  错误: 解析开始时间失败: %v\n", err)
        return false
    }
    closeHour, closeMin, err := parseTime(restaurant.CloseTime)
    if err != nil {
        fmt.Printf("  错误: 解析结束时间失败: %v\n", err)
        return false
    }
    
    fmt.Printf("  调试: 营业时间解析 - 开始: %02d:%02d, 结束: %02d:%02d\n", openHour, openMin, closeHour, closeMin)
    
    // Get the date and time components
    startDate := start.Truncate(24 * time.Hour)
    endDate := end.Truncate(24 * time.Hour)
    
    fmt.Printf("  调试: 日期范围 - %s 到 %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
    
    // Check each day of the reservation
    for date := startDate; !date.After(endDate); date = date.Add(24 * time.Hour) {
        fmt.Printf("  调试: 检查日期 %s\n", date.Format("2006-01-02"))
        
        openTime := date.Add(time.Duration(openHour)*time.Hour + time.Duration(openMin)*time.Minute)
        closeTime := date.Add(time.Duration(closeHour)*time.Hour + time.Duration(closeMin)*time.Minute)
        
        fmt.Printf("  调试: 当日营业时间 %s - %s\n", openTime.Format("15:04"), closeTime.Format("15:04"))
        
        // Handle overnight hours (e.g., 22:00 - 02:00)
        if closeTime.Before(openTime) {
            closeTime = closeTime.Add(24 * time.Hour)
            fmt.Printf("  调试: 跨夜营业，关闭时间调整为 %s\n", closeTime.Format("15:04"))
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
        
        fmt.Printf("  调试: 当日预订时间 %s - %s\n", dayStart.Format("15:04"), dayEnd.Format("15:04"))
        
        // If there's any part of the reservation on this day
        if dayStart.Before(dayEnd) {
            fmt.Printf("  调试: 该日有预订时间段\n")
            // Check if this part is within operating hours
            if dayStart.Before(openTime) || dayEnd.After(closeTime) {
                fmt.Printf("  调试: 预订时间超出营业时间！开始:%s < 营业开始:%s? %t | 结束:%s > 营业结束:%s? %t\n", 
                    dayStart.Format("15:04"), openTime.Format("15:04"), dayStart.Before(openTime),
                    dayEnd.Format("15:04"), closeTime.Format("15:04"), dayEnd.After(closeTime))
                return false // Any part outside operating hours means rejection
            }
            fmt.Printf("  调试: 该日预订时间在营业时间内\n")
        } else {
            fmt.Printf("  调试: 该日无预订时间段\n")
        }
    }
    
    fmt.Printf("  调试: 所有时间段验证通过\n")
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