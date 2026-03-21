# admin-demo-go

基于 `Gin + GORM + Redis + NSQ` 的后台管理系统后端基础骨架。

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
- `internal/admin`：后台通用管理模块（用户/角色/菜单）
- `internal/handler`：HTTP 接口
- `internal/service`：业务逻辑
- `internal/repository`：数据访问
- `docs/openapi`：OpenAPI 接口定义
- `sql`：数据库初始化脚本

## 配置

复制并修改 `configs/config.example.yaml`。  
按你的要求，MySQL/Redis/NSQ 连接信息当前均为空，后续补齐即可。

默认数据库名字段为 `admin_demo`。

文件资源存储支持三种模式（仅改配置，不改前端交互）：

- `storage.mode: local` 本地磁盘
- `storage.mode: aws-s3` AWS S3
- `storage.mode: minio` 自建 MinIO

## 启动

```bash
go mod tidy
go run ./cmd/server
```
