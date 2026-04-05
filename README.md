# admin-demo-go

基于 `Gin + GORM + Redis + NSQ` 的后台管理系统后端基础骨架，默认使用 MySQL，适合作为新后台项目的可复用后端基座。

## 已实现基础能力

- 登录（含谷歌验证码字段 `googleCode`，演示默认验证码 `123456`）
- 注册
- 个人信息查询与更新
- 分级权限返回（`admin/editor/test`）
- 菜单权限接口示例（`/api/menu/navigate`）
- Google Authenticator 真实绑定流程（生成二维码、绑定、解绑）
- RBAC 持久化表结构（用户/角色/权限/菜单及关系表）
- 统一错误码与请求追踪（`request_id`）
- 访问日志 + 业务事件 NSQ 投递（未配置 NSQ 时自动降级为本地日志）

## 新增接口

- `GET /api/auth/google/setup` 生成绑定二维码
- `POST /api/auth/google/bind` 提交验证码完成绑定
- `POST /api/auth/google/unbind` 验证码解绑
- `GET /api/rbac/permissions` 获取当前用户权限码
- `GET /api/rbac/menus` 获取当前用户菜单

## 目录说明

- `cmd/server`：服务入口
- `internal/config`：配置加载
- `internal/bootstrap`：DB/Redis/NSQ 初始化
- `internal/admin`：后台通用管理模块（用户/角色/菜单/字典/系统配置/操作日志）
- `internal/handler`：HTTP 接口
- `internal/service`：业务逻辑
- `internal/repository`：数据访问
- `docs/openapi`：OpenAPI 接口定义
- `sql`：数据库初始化脚本

## 配置

默认读取 `configs/config.example.yaml`，也可以通过环境变量 `ADMIN_DEMO_CONFIG` 指定配置文件。

当前默认配置策略：

- `mysql.dsn` 必填，服务启动时必须连接 MySQL
- `redis.addr` 为空时，跳过 Redis 初始化
- `nsq.producer_addr` 为空时，跳过 NSQ 初始化
- `storage.mode=local` 时，文件默认存到 `storage/`
- 文件存储模式仅允许通过服务端配置文件切换，不通过前端页面切换

文件资源存储支持三种模式（仅改配置，不改前端交互）：

- `storage.mode: local` 本地磁盘
- `storage.mode: aws-s3` AWS S3
- `storage.mode: minio` 自建 MinIO

## 启动

```bash
go mod tidy
go run ./cmd/server
```

默认端口：`8889`

健康检查：

```bash
curl http://127.0.0.1:8889/healthz
```

## 默认演示账号

- 管理员：`admin / 123456`
- 编辑员：`editor / 123456`
- 测试员：`test / 123456`

说明：

- 首次启动会自动建表并写入演示数据
- Google 二次验证演示验证码默认为 `123456`
