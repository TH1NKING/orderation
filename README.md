# 餐饮预订系统后端（Go）

一个基于 Go 标准库实现的 Web 后端原型，用于《基于 Web 的餐饮预订系统设计与实现》。

- 无外部依赖（仅使用 `net/http` 等标准库）。
- 内存存储（重启数据会丢失），提供接口抽象，便于后续切换为数据库。
- 简单的基于 HMAC 的 Token（类 JWT）认证与授权；密码哈希使用迭代加盐的 SHA-256（教学/原型用途）。

## 运行

```bash
# 进入项目根目录
go run ./cmd/server
```

环境变量：
- `ADDR`：监听地址，默认 `:8080`
- `SECRET`：签名密钥（HMAC），默认随机生成（重启后 token 失效）
- `ADMIN_EMAIL`、`ADMIN_PASSWORD`、`ADMIN_NAME`：启动时自动创建管理员账号（可选）

## 目录结构

- `cmd/server/main.go`：入口
- `internal/server`：服务组装与路由挂载
- `internal/web/router`：极简路由（支持 `:id` 路径参数）
- `internal/web/middleware`：认证/鉴权中间件
- `internal/web/handlers`：各 API 处理器
- `internal/models`：模型定义
- `internal/store`：存储接口
- `internal/store/memory`：内存存储实现
- `internal/auth`：密码哈希与 Token

## API 概览

所有请求与响应均为 JSON。时间使用 RFC3339（例如：`2025-09-07T18:00:00+08:00`）。

- 健康检查
  - GET `/healthz` → `{ "ok": true }`

- 认证
  - POST `/api/v1/auth/register`
    - 请求：`{ "name": "张三", "email": "a@b.com", "password": "123456" }`
    - 响应：`{ "token": "...", "user": {"id":"...","name":"张三","email":"a@b.com","role":"user"} }`
  - POST `/api/v1/auth/login`
    - 请求：`{ "email": "a@b.com", "password": "123456" }`
    - 响应同上
  - 认证方式：`Authorization: Bearer <token>`

- 餐厅（管理员）
  - POST `/api/v1/restaurants`（需 admin）
    - 请求：`{ "name":"店名","address":"地址","openTime":"10:00","closeTime":"22:00" }`
    - 响应：`Restaurant`
  - GET `/api/v1/restaurants` → `Restaurant[]`
  - GET `/api/v1/restaurants/:id` → `Restaurant`

- 餐桌（管理员）
  - POST `/api/v1/restaurants/:id/tables`（需 admin）
    - 请求：`{ "name":"A1","capacity":4 }`
    - 响应：`Table`
  - GET `/api/v1/restaurants/:id/tables?minCapacity=4` → `Table[]`

- 余位查询与预订
  - POST `/api/v1/restaurants/:id/availability`
    - 请求：`{ "start":"2025-09-07T18:00:00+08:00", "end":"2025-09-07T20:00:00+08:00", "guests": 2 }`
    - 响应：`[{ "tableId":"...","capacity":4 }]`
  - POST `/api/v1/restaurants/:id/reservations`（需登录）
    - 请求：
      - 自动分配餐桌：`{ "start":"...","end":"...","guests":2 }`
      - 指定餐桌：`{ "start":"...","end":"...","guests":2, "tableId":"..." }`
    - 响应：`Reservation`
  - DELETE `/api/v1/reservations/:id`（预订者或 admin）
    - 响应：`{ "status":"cancelled" }`
  - GET `/api/v1/me/reservations`（需登录） → `Reservation[]`

### 数据模型（简化）

- `User`：`{ id, name, email, role, createdAt }`
- `Restaurant`：`{ id, name, address, openTime, closeTime, createdAt }`
- `Table`：`{ id, restaurantId, name, capacity, createdAt }`
- `Reservation`：`{ id, restaurantId, tableId, userId, startTime, endTime, guests, status, createdAt }`

## 实现说明与后续建议

- 当前使用内存存储（`sync.RWMutex` 并发安全）；可替换为数据库：
  - 定义 `internal/store` 的接口实现（如 Postgres/MySQL/SQLite）。
  - 在 `internal/server/server.go` 替换为具体实现即可。
- 密码哈希与 Token 仅用于教学原型：
  - 生产建议改为 `bcrypt/scrypt/argon2` 与标准 JWT 库。
- 可拓展功能：
  - 复杂营业时间与跨天预订、合桌/并桌策略、超时释放等。
  - 预订修改、订单/支付、通知（短信/邮件）。
  - 分页与筛选、审计日志、限流与安全审计。

## 快速测试顺序（示例）

1) 启动并创建管理员（通过环境变量）
```bash
ADMIN_EMAIL=admin@demo.com ADMIN_PASSWORD=admin go run ./cmd/server
```
2) 创建餐厅（用管理员 token）
3) 创建餐桌
4) 注册/登录普通用户
5) 查询余位并下单
6) 查看/取消我的预订
```
