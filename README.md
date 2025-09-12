# Orderation - 餐厅预订管理系统

一个功能完整的餐厅预订管理系统，使用 Go 语言后端和原生 JavaScript 前端构建。支持用户管理、餐厅管理、桌台管理和智能预订系统。

## ✨ 功能特性

### 🔐 用户系统

- **用户注册和登录** - JWT token 认证
- **角色权限管理** - 普通用户和管理员权限区分
- **密码安全** - 使用 bcrypt 加密存储

### 🏪 餐厅管理

- **餐厅信息管理** - 创建、查看、删除餐厅（管理员功能）
- **营业时间设置** - 支持自定义营业时间
- **餐厅列表浏览** - 所有用户可查看餐厅信息

### 🪑 桌台管理

- **桌台信息管理** - 创建桌台，设置容量（管理员功能）
- **桌台状态查看** - 实时显示桌台占用状态
- **按餐厅分组** - 每个餐厅独立管理桌台

### 📅 预订系统

- **智能预订** - 自动匹配最适合的桌台
- **时间验证** - 确保预订时间在营业时间内
- **冲突检测** - 防止重复预订同一桌台
- **预订管理** - 用户可查看、取消自己的预订

### 🔍 可用性查询

- **实时查询** - 查询指定时间段的可用桌台
- **智能推荐** - 根据人数推荐合适容量的桌台
- **快速预订** - 直接从可用性查询结果快速预订

## 🏗️ 技术架构

### 后端（Go）

- **框架**: 原生 `net/http` + 自定义路由器
- **数据库**: MySQL（主要）+ 内存存储（fallback）
- **认证**: JWT + bcrypt 密码哈希
- **架构模式**: 分层架构（Handler → Service → Store）

### 前端

- **技术栈**: 原生 HTML + CSS + JavaScript
- **样式**: 响应式设计，支持移动端
- **交互**: 基于 fetch API 的异步请求
- **状态管理**: localStorage 存储用户会话

## 🚀 快速开始

### 环境要求

- Go 1.22.0+
- MySQL 5.7+ （可选）
- 现代浏览器

### 方式一：直接运行

```bash
# 克隆项目
git clone <repository-url>
cd orderation

# 安装依赖
go mod tidy

# 启动服务器
go run cmd/server/main.go
```

### 环境变量配置

- `ADDR`：监听地址，默认 `:8080`
- `SECRET`：JWT 签名密钥，默认随机生成
- `ADMIN_EMAIL`、`ADMIN_PASSWORD`、`ADMIN_NAME`：启动时自动创建管理员账号
- `MYSQL_HOST`、`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_DATABASE`：MySQL 数据库配置

### 方式二：使用 Docker（推荐）

项目提供 `docker-compose.yml` 配置，包含：

- MySQL 8.0 数据库（持久化存储）
- 应用容器（自动建表）

```bash
# 复制环境变量配置
cp .env.example .env

# 启动服务（后台运行）
docker-compose up -d --build

# 访问应用
open http://localhost:8080
```

停止服务：

```bash
docker-compose down

# 清理数据（可选）
docker-compose down -v
```

### 数据库初始化

如果使用 MySQL，可以运行初始化脚本来创建示例数据：

```bash
# 创建示例餐厅、桌台和用户
go run scripts/init_db.go

# 检查用户账户
go run check_users.go
```

### 访问应用

启动后，在浏览器中打开 `http://localhost:8080` 即可访问前端界面。

### 默认管理员账户

系统已创建测试管理员账户：

- **邮箱**: `simpleadmin@test.com`
- **密码**: `123456`
- **权限**: 创建/删除餐厅、创建桌台

## 📁 项目结构

```
orderation/
├── cmd/server/           # 服务器入口
│   └── main.go
├── internal/
│   ├── auth/            # 认证模块（JWT + 密码哈希）
│   ├── models/          # 数据模型定义
│   ├── server/          # 服务器配置和初始化
│   ├── store/           # 数据存储层接口
│   │   ├── mysql/       # MySQL 存储实现
│   │   └── memory/      # 内存存储实现
│   └── web/
│       ├── handlers/    # HTTP 请求处理器
│       ├── middleware/  # 认证和权限中间件
│       └── router/      # 自定义路由器
├── web/                 # 前端文件
│   ├── index.html       # 主页面
│   └── app.js          # JavaScript 逻辑
├── scripts/             # 工具脚本
│   └── init_db.go      # 数据库初始化
├── docker-compose.yml   # Docker 配置
└── .env.example        # 环境变量示例
```

## 🔧 API 接口文档

所有请求与响应均为 JSON 格式。认证采用 JWT Bearer Token。

### 认证接口

```http
POST /api/v1/auth/register  # 用户注册
POST /api/v1/auth/login     # 用户登录
```

### 餐厅接口

```http
GET    /api/v1/restaurants           # 获取餐厅列表
GET    /api/v1/restaurants/:id       # 获取餐厅详情  
GET    /api/v1/restaurants/:id/details # 获取餐厅详细信息
POST   /api/v1/restaurants           # 创建餐厅（管理员）
DELETE /api/v1/restaurants/:id       # 删除餐厅（管理员）
```

### 桌台接口

```http
GET  /api/v1/restaurants/:id/tables  # 获取餐厅桌台列表
POST /api/v1/restaurants/:id/tables  # 创建桌台（管理员）
```

### 预订接口

```http
POST   /api/v1/restaurants/:id/availability  # 查询可用桌台
POST   /api/v1/restaurants/:id/reservations   # 创建预订（需登录）
GET    /api/v1/me/reservations               # 查看我的预订（需登录）
DELETE /api/v1/reservations/:id             # 取消预订（需登录）
```

### 请求示例

#### 用户注册

```json
POST /api/v1/auth/register
{
  "name": "张三",
  "email": "zhangsan@example.com", 
  "password": "123456"
}
```

#### 创建餐厅（管理员）

```json
POST /api/v1/restaurants
Authorization: Bearer <admin-token>
{
  "name": "川菜馆",
  "address": "北京市朝阳区xxx街道123号",
  "openTime": "09:00",
  "closeTime": "22:00"
}
```

#### 查询可用桌台

```json
POST /api/v1/restaurants/:id/availability
{
  "start": "2025-01-15T18:00:00+08:00",
  "end": "2025-01-15T20:00:00+08:00",
  "guests": 4
}
```

## 🛡️ 安全特性

- **JWT 认证** - 无状态的安全令牌认证
- **角色权限控制** - 基于角色的访问控制（RBAC）
- **密码加密** - bcrypt 哈希存储
- **CORS 支持** - 跨域请求处理
- **输入验证** - 服务端数据验证和清理
- **SQL 注入防护** - 参数化查询

## 🔍 业务逻辑亮点

### 智能桌台分配

- 优先分配容量最接近需求的桌台
- 避免大桌台被小团体占用
- 自动处理容量冲突

### 营业时间验证

- 支持跨时区时间处理（Asia/Shanghai）
- 精确到分钟的营业时间控制
- 防止非营业时间预订

### 冲突检测

- 实时检测时间段重叠
- 支持复杂的预订时间验证
- 确保数据一致性

## 🧪 开发工具

项目包含多个调试和管理工具：

```bash
# 检查数据库用户
go run check_users.go

# 调试可用性查询
go run debug_availability.go  

# 调试营业时间验证
go run debug_hours.go
```

## 📊 数据模型

### User（用户）

```json
{
  "id": "uuid",
  "name": "用户名",
  "email": "邮箱地址",
  "role": "user|admin",
  "createdAt": "2025-01-15T10:00:00Z"
}
```

### Restaurant（餐厅）

```json
{
  "id": "uuid",
  "name": "餐厅名称",
  "address": "餐厅地址", 
  "openTime": "09:00",
  "closeTime": "22:00",
  "createdAt": "2025-01-15T10:00:00Z"
}
```

### Table（桌台）

```json
{
  "id": "uuid",
  "restaurantId": "餐厅ID",
  "name": "桌台名称",
  "capacity": 4,
  "createdAt": "2025-01-15T10:00:00Z"
}
```

### Reservation（预订）

```json
{
  "id": "uuid",
  "restaurantId": "餐厅ID",
  "tableId": "桌台ID",
  "userId": "用户ID", 
  "startTime": "2025-01-15T18:00:00+08:00",
  "endTime": "2025-01-15T20:00:00+08:00",
  "guests": 4,
  "status": "confirmed|pending|cancelled|completed",
  "createdAt": "2025-01-15T10:00:00Z"
}
```

## 🚧 待改进功能

- [ ] 预订提醒系统（短信/邮件）
- [ ] 餐厅评价和评分功能
- [ ] 移动端 PWA 应用
- [ ] 数据统计和分析面板
- [ ] 多语言国际化支持
- [ ] 第三方支付集成
- [ ] 预订修改功能
- [ ] 合桌和并桌策略
- [ ] API 限流和缓存
- [ ] 审计日志系统

## ❓ 常见问题

**Q: 端口 8080 已被占用怎么办？**
A: 修改 `.env` 文件中的 `ADDR` 变量，如 `ADDR=:8081`

**Q: MySQL 连接失败？**
A: 确认数据库配置正确，可以先运行 `docker-compose up -d mysql` 启动数据库

**Q: Token 登录失效？**
A: 重启服务器后 JWT 密钥会重新生成，需要重新登录

**Q: 前端界面显示异常？**
A: 检查浏览器开发者工具的 Console 和 Network 标签页，确认 API 请求正常

**Q: 管理员功能不显示？**
A: 确认使用管理员账户登录（`simpleadmin@test.com` / `123456`）

## 🤝 贡献指南

1. Fork 项目到你的账户
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发规范

- 遵循 Go 语言标准代码规范
- 添加必要的单元测试
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 📞 支持与反馈

如有问题或建议，欢迎：

- 提交 [Issue](../../issues)
- 发起 [Pull Request](../../pulls)

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！

💡 **快速体验流程**：

1. 启动服务：`docker-compose up -d`
2. 访问页面：`http://localhost:8080`
3. 管理员登录：`simpleadmin@test.com` / `123456`
4. 创建餐厅和桌台
5. 注册普通用户体验预订功能
