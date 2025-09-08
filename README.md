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
- `MYSQL_DSN`：若设置则启用 MySQL 存储（例如：`app:appl3pass@tcp(127.0.0.1:3306)/orderation?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci`），未设置则使用内存存储

### 使用 Docker 一键部署（含 MySQL）

已提供 `Dockerfile` 与 `docker-compose.yml`，包含：
- `mysql:8.0` 数据库（持久化到 `db_data` 卷）
- 应用容器（启动时自动创建表，无需手动迁移）

启动：

```bash
cp .env.example .env   # 可选：创建并修改你的环境变量
docker compose up -d --build
# 等待 db 健康检查通过后，app 将自动启动，默认 http://localhost:8080
```

环境变量（.env 自动加载，可按需修改）：
- `MYSQL_ROOT_PASSWORD`、`MYSQL_DATABASE`、`MYSQL_USER`、`MYSQL_PASSWORD`
- `MYSQL_DSN`: `app:appl3pass@tcp(db:3306)/orderation?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci`
- `SECRET`: 示例值 `change-me-in-prod`（生产请修改）
- `ADMIN_EMAIL`/`ADMIN_PASSWORD`: 首次启动自动创建管理员

停止并清理：

```bash
docker compose down
# 如需删除数据卷：
docker compose down -v
```

### 本地开发（连接 MySQL）

你可以只运行数据库容器，本地 `go run` 应用：

```bash
# 启动 MySQL（后台）
docker compose up -d db

# 设置 DSN 并运行应用（连接到宿主机上的 MySQL）
export MYSQL_DSN="app:appl3pass@tcp(127.0.0.1:3306)/orderation?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
export SECRET="dev-secret" # 本地开发示例
export ADMIN_EMAIL="admin@demo.local"
export ADMIN_PASSWORD="admin123"
go run ./cmd/server
```

DSN 速查：
- 在容器内连接 compose 中的 MySQL：`app:appl3pass@tcp(db:3306)/orderation?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci`
- 在宿主机连接 compose 中的 MySQL：`app:appl3pass@tcp(127.0.0.1:3306)/orderation?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci`

提示：Docker Compose 会自动从 `docker-compose.yml` 同目录下加载 `.env` 文件用于变量替换；本仓库提供 `.env.example`，可复制为 `.env` 并按需修改。

## 目录结构

- `cmd/server/main.go`：入口
- `internal/server`：服务组装与路由挂载
- `internal/web/router`：极简路由（支持 `:id` 路径参数）
- `internal/web/middleware`：认证/鉴权中间件
- `internal/web/handlers`：各 API 处理器
- `internal/models`：模型定义
- `internal/store`：存储接口
- `internal/store/memory`：内存存储实现
- `internal/store/mysql`：MySQL 存储实现与自动建表
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

## API 示例（curl）

以下命令可在 Docker 启动后直接在终端执行：

```bash
BASE=http://localhost:8080

# 1) 管理员登录（使用 compose 中的环境变量）
ADMIN_TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"admin123"}' | jq -r .token)
echo "ADMIN_TOKEN=$ADMIN_TOKEN"

# 2) 创建餐厅（需要管理员）
REST=$(curl -s -X POST "$BASE/api/v1/restaurants" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"示例餐厅","address":"地址","openTime":"10:00","closeTime":"22:00"}')
REST_ID=$(echo "$REST" | jq -r .id)
echo "REST_ID=$REST_ID"

# 3) 创建餐桌（需要管理员）
TBL=$(curl -s -X POST "$BASE/api/v1/restaurants/$REST_ID/tables" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"A1","capacity":4}')
TABLE_ID=$(echo "$TBL" | jq -r .id)
echo "TABLE_ID=$TABLE_ID"

# 4) 注册普通用户
REG=$(curl -s -X POST "$BASE/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d '{"name":"张三","email":"u@test.local","password":"p"}')
USER_TOKEN=$(echo "$REG" | jq -r .token)
echo "USER_TOKEN=$USER_TOKEN"

# 5) 查询余位（时间示例：2 小时后到 4 小时后）
START=$(date -u -v+2H +"%Y-%m-%dT%H:00:00Z" 2>/dev/null || date -u -d "+2 hours" +"%Y-%m-%dT%H:00:00Z")
END=$(date -u -v+4H +"%Y-%m-%dT%H:00:00Z" 2>/dev/null || date -u -d "+4 hours" +"%Y-%m-%dT%H:00:00Z")
curl -s -X POST "$BASE/api/v1/restaurants/$REST_ID/availability" -H 'Content-Type: application/json' \
  -d "{\"start\":\"$START\",\"end\":\"$END\",\"guests\":2}"

# 6) 创建预订（指定餐桌）
RES=$(curl -s -X POST "$BASE/api/v1/restaurants/$REST_ID/reservations" \
  -H "Authorization: Bearer $USER_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"start\":\"$START\",\"end\":\"$END\",\"guests\":2,\"tableId\":\"$TABLE_ID\"}")
RES_ID=$(echo "$RES" | jq -r .id)
echo "RES_ID=$RES_ID"

# 7) 查看我的预订
curl -s -H "Authorization: Bearer $USER_TOKEN" "$BASE/api/v1/me/reservations" | jq .

# 8) 取消预订
curl -s -X DELETE -H "Authorization: Bearer $USER_TOKEN" "$BASE/api/v1/reservations/$RES_ID" | jq .
```

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

## 常见问题

- 端口占用：修改 `ADDR` 或关闭占用该端口的进程。
- MySQL 连接失败：确认 `MYSQL_DSN` 正确、数据库已就绪（compose 启动有健康检查）。
- Token 失效：未设置 `SECRET` 时会使用随机密钥，重启后历史 token 失效；生产务必显式设置。

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
