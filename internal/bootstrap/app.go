package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"admin-demo-go/internal/config"
	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/event"
	"admin-demo-go/internal/storage"

	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type App struct {
	Cfg         *config.Config
	DB          *gorm.DB
	Redis       *redis.Client
	NSQProducer *nsq.Producer
	EventPub    *event.Publisher
	Storage     storage.Client
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{Cfg: cfg}

	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("init mysql failed: %w", err)
	}
	app.DB = db

	rdb, err := initRedis(cfg)
	if err != nil {
		log.Printf("init redis skipped: %v", err)
	} else {
		app.Redis = rdb
	}

	producer, err := initNSQ(cfg)
	if err != nil {
		log.Printf("init nsq skipped: %v", err)
	} else {
		app.NSQProducer = producer
	}
	app.EventPub = event.NewPublisher(app.NSQProducer, cfg.NSQ.Topic)
	st, err := initStorage(cfg)
	if err != nil {
		log.Printf("init storage failed, fallback local: %v", err)
		st = storage.NewLocal(cfg.Storage.Local.BaseDir)
	}
	app.Storage = st

	return app, nil
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	if cfg.MySQL.DSN == "" {
		return nil, fmt.Errorf("mysql dsn is empty")
	}
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MySQL.ConnMaxLifetimeMinutes) * time.Minute)
	if err = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.Menu{},
		&model.DictItem{},
		&model.SystemConfig{},
		&model.OperationLog{},
		&model.Department{},
		&model.Notice{},
		&model.NoticeRead{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.RoleMenu{},
	); err != nil {
		return nil, err
	}
	seedDemoData(db)
	ensureDemoUsers(db, cfg.App.Mode)
	return db, nil
}

func seedDemoData(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		seedRBAC(db)
		return
	}
	db.Create(&model.User{
		Username:     "admin",
		PasswordHash: "$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy", // 123456
		Role:         "admin",
		Nickname:     "系统管理员",
		Avatar:       "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png",
		GoogleSecret: "",
	})
	db.Create(&model.User{
		Username:     "editor",
		PasswordHash: "$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy",
		Role:         "editor",
		Nickname:     "内容编辑",
		Avatar:       "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_2.png",
		GoogleSecret: "",
	})
	db.Create(&model.User{
		Username:     "test",
		PasswordHash: "$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy",
		Role:         "test",
		Nickname:     "测试账号",
		Avatar:       "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_3.png",
		GoogleSecret: "",
	})
	seedRBAC(db)

	var users []model.User
	var roles []model.Role
	db.Find(&users)
	db.Find(&roles)
	roleMap := map[string]uint{}
	for _, r := range roles {
		roleMap[r.Code] = r.ID
	}
	for _, u := range users {
		code := u.Role
		if code == "" {
			code = "editor"
		}
		if rid, ok := roleMap[code]; ok {
			db.FirstOrCreate(&model.UserRole{}, model.UserRole{
				UserID: u.ID,
				RoleID: rid,
			})
		}
	}
}

func seedRBAC(db *gorm.DB) {
	roles := []model.Role{
		{Code: "admin", Name: "管理员"},
		{Code: "editor", Name: "编辑员"},
		{Code: "test", Name: "测试员"},
	}
	for _, r := range roles {
		db.FirstOrCreate(&model.Role{}, model.Role{Code: r.Code, Name: r.Name})
	}

	perms := []model.Permission{
		{Code: "dashboard:view", Name: "查看仪表盘", Type: "menu"},
		{Code: "profile:update", Name: "更新个人信息", Type: "api"},
		{Code: "google:bind", Name: "绑定谷歌验证", Type: "api"},
		{Code: "rbac:view", Name: "查看权限信息", Type: "api"},
		{Code: "permission:switch", Name: "权限切换按钮", Type: "button"},
	}
	for _, p := range perms {
		db.FirstOrCreate(&model.Permission{}, model.Permission{Code: p.Code, Name: p.Name, Type: p.Type})
	}

	type menuSeed struct {
		Key    string
		Parent string
		Menu   model.Menu
	}
	menuDefs := []menuSeed{
		{Key: "Root", Parent: "", Menu: model.Menu{Path: "/", Name: "Root", Component: "Layout", Redirect: "index", Title: "首页", Icon: "home", Sort: 1}},
		{Key: "Index", Parent: "Root", Menu: model.Menu{Path: "index", Name: "Index", Component: "@/views/index/index", Title: "首页", Icon: "home", Affix: true, Sort: 1}},
		{Key: "Vab", Parent: "", Menu: model.Menu{Path: "/vab", Name: "Vab", Component: "Layout", Redirect: "noRedirect", Title: "组件", Icon: "box-open", AlwaysShow: true, Sort: 10}},
		{Key: "Permission", Parent: "Vab", Menu: model.Menu{Path: "permissions", Name: "Permission", Component: "@/views/vab/permissions/index", Title: "角色权限", PermissionCode: "rbac:view", Sort: 11}},
		{Key: "PersonnelManagement", Parent: "", Menu: model.Menu{Path: "/personnelManagement", Name: "PersonnelManagement", Component: "Layout", Redirect: "noRedirect", Title: "配置", Icon: "users-cog", Sort: 20}},
		{Key: "RoleManagement", Parent: "PersonnelManagement", Menu: model.Menu{Path: "roleManagement", Name: "RoleManagement", Component: "@/views/personnelManagement/roleManagement/index", Title: "角色管理", PermissionCode: "rbac:view", Sort: 21}},
		{Key: "DepartmentManagement", Parent: "PersonnelManagement", Menu: model.Menu{Path: "departmentManagement", Name: "DepartmentManagement", Component: "@/views/personnelManagement/departmentManagement/index", Title: "部门管理", PermissionCode: "rbac:view", Sort: 22}},
		{Key: "NoticeManagement", Parent: "PersonnelManagement", Menu: model.Menu{Path: "noticeManagement", Name: "NoticeManagement", Component: "@/views/personnelManagement/noticeManagement/index", Title: "通知公告", PermissionCode: "rbac:view", Sort: 23}},
		{Key: "DictManagement", Parent: "PersonnelManagement", Menu: model.Menu{Path: "dictManagement", Name: "DictManagement", Component: "@/views/personnelManagement/dictManagement/index", Title: "字典管理", PermissionCode: "rbac:view", Sort: 24}},
		{Key: "ConfigManagement", Parent: "PersonnelManagement", Menu: model.Menu{Path: "configManagement", Name: "ConfigManagement", Component: "@/views/personnelManagement/configManagement/index", Title: "系统配置", PermissionCode: "rbac:view", Sort: 25}},
		{Key: "OperationLog", Parent: "PersonnelManagement", Menu: model.Menu{Path: "operationLog", Name: "OperationLog", Component: "@/views/personnelManagement/operationLog/index", Title: "操作日志", PermissionCode: "rbac:view", Sort: 26}},
	}
	menuPK := map[string]uint{}
	for _, def := range menuDefs {
		var menu model.Menu
		db.Where("name = ? and path = ?", def.Menu.Name, def.Menu.Path).FirstOrCreate(&menu, def.Menu)
		menuPK[def.Key] = menu.ID
	}
	for _, def := range menuDefs {
		parentID := uint(0)
		if def.Parent != "" {
			parentID = menuPK[def.Parent]
		}
		def.Menu.ParentID = parentID
		db.Model(&model.Menu{}).Where("id = ?", menuPK[def.Key]).Updates(def.Menu)
	}

	var roleList []model.Role
	var permList []model.Permission
	var menuList []model.Menu
	db.Find(&roleList)
	db.Find(&permList)
	db.Find(&menuList)
	roleID := map[string]uint{}
	permID := map[string]uint{}
	menuID := map[string]uint{}
	for _, r := range roleList {
		roleID[r.Code] = r.ID
	}
	for _, p := range permList {
		permID[p.Code] = p.ID
	}
	for _, m := range menuList {
		menuID[m.Name] = m.ID
	}

	adminPerms := []string{"dashboard:view", "profile:update", "google:bind", "rbac:view", "permission:switch"}
	editorPerms := []string{"dashboard:view", "profile:update", "google:bind", "rbac:view"}
	testPerms := []string{"dashboard:view", "profile:update", "google:bind"}
	bindRolePerms(db, roleID["admin"], adminPerms, permID)
	bindRolePerms(db, roleID["editor"], editorPerms, permID)
	bindRolePerms(db, roleID["test"], testPerms, permID)

	adminMenus := []string{"Root", "Index", "Vab", "Permission", "PersonnelManagement", "RoleManagement", "DepartmentManagement", "NoticeManagement", "DictManagement", "ConfigManagement", "OperationLog"}
	editorMenus := []string{"Root", "Index", "Vab", "Permission"}
	testMenus := []string{"Root", "Index"}
	bindRoleMenus(db, roleID["admin"], adminMenus, menuID)
	bindRoleMenus(db, roleID["editor"], editorMenus, menuID)
	bindRoleMenus(db, roleID["test"], testMenus, menuID)

	seedDictItems(db)
	seedSystemConfigs(db)
	seedDepartments(db)
	seedNotices(db)
}

func bindRolePerms(db *gorm.DB, roleID uint, codes []string, permIDMap map[string]uint) {
	for _, code := range codes {
		if pid, ok := permIDMap[code]; ok {
			db.FirstOrCreate(&model.RolePermission{}, model.RolePermission{RoleID: roleID, PermissionID: pid})
		}
	}
}

func bindRoleMenus(db *gorm.DB, roleID uint, names []string, menuIDMap map[string]uint) {
	for _, name := range names {
		if mid, ok := menuIDMap[name]; ok {
			db.FirstOrCreate(&model.RoleMenu{}, model.RoleMenu{RoleID: roleID, MenuID: mid})
		}
	}
}

func seedDictItems(db *gorm.DB) {
	items := []model.DictItem{
		{DictType: "user_status", Label: "启用", Value: "enabled", Status: "enabled", Sort: 1, Remark: "用户状态"},
		{DictType: "user_status", Label: "禁用", Value: "disabled", Status: "enabled", Sort: 2, Remark: "用户状态"},
		{DictType: "notice_level", Label: "普通", Value: "normal", Status: "enabled", Sort: 1, Remark: "通知等级"},
		{DictType: "notice_level", Label: "重要", Value: "important", Status: "enabled", Sort: 2, Remark: "通知等级"},
	}
	for _, item := range items {
		db.Where("dict_type = ? AND value = ?", item.DictType, item.Value).FirstOrCreate(&model.DictItem{}, item)
	}
}

func seedSystemConfigs(db *gorm.DB) {
	items := []model.SystemConfig{
		{ConfigKey: "site.title", ConfigValue: "Admin Demo", Name: "站点标题", Group: "site", ValueType: "string", Remark: "后台系统标题"},
		{ConfigKey: "site.description", ConfigValue: "可复用后台管理系统基座", Name: "站点描述", Group: "site", ValueType: "string", Remark: "站点描述文案"},
		{ConfigKey: "site.logo", ConfigValue: "/logo.png", Name: "站点 Logo", Group: "site", ValueType: "string", Remark: "站点 logo 地址"},
		{ConfigKey: "security.login_captcha", ConfigValue: "false", Name: "登录验证码", Group: "security", ValueType: "boolean", Remark: "是否启用登录验证码"},
		{ConfigKey: "security.password_min_length", ConfigValue: "8", Name: "密码最小长度", Group: "security", ValueType: "number", Remark: "密码安全策略"},
		{ConfigKey: "security.two_factor_auth", ConfigValue: "false", Name: "双重认证", Group: "security", ValueType: "boolean", Remark: "是否要求双重认证"},
		{ConfigKey: "security.session_timeout", ConfigValue: "30", Name: "会话超时", Group: "security", ValueType: "number", Remark: "单位：分钟"},
		{ConfigKey: "security.max_login_attempts", ConfigValue: "5", Name: "登录失败次数", Group: "security", ValueType: "number", Remark: "超过次数可触发锁定策略"},
		{ConfigKey: "email.smtp_server", ConfigValue: "smtp.example.com", Name: "SMTP服务器", Group: "email", ValueType: "string", Remark: "SMTP 地址"},
		{ConfigKey: "email.smtp_port", ConfigValue: "587", Name: "SMTP端口", Group: "email", ValueType: "number", Remark: "SMTP 端口"},
		{ConfigKey: "email.username", ConfigValue: "user@example.com", Name: "用户名", Group: "email", ValueType: "string", Remark: "发信账号"},
		{ConfigKey: "email.password", ConfigValue: "", Name: "密码", Group: "email", ValueType: "string", Remark: "发信密码"},
		{ConfigKey: "email.sender_email", ConfigValue: "noreply@example.com", Name: "发件人邮箱", Group: "email", ValueType: "string", Remark: "默认发件邮箱"},
	}
	for _, item := range items {
		db.Where("config_key = ?", item.ConfigKey).FirstOrCreate(&model.SystemConfig{}, item)
	}
}

func seedDepartments(db *gorm.DB) {
	type deptSeed struct {
		ParentCode string
		Item       model.Department
	}
	items := []deptSeed{
		{Item: model.Department{Name: "总部", Code: "headquarters", Leader: "张总", Phone: "13800000000", Status: "enabled", Sort: 1, Remark: "公司总部"}},
		{Item: model.Department{Name: "研发中心", Code: "rd-center", Leader: "李工", Phone: "13800000001", Status: "enabled", Sort: 2, Remark: "产品与研发"}},
		{Item: model.Department{Name: "运营中心", Code: "ops-center", Leader: "王运", Phone: "13800000002", Status: "enabled", Sort: 3, Remark: "内容与增长"}},
		{ParentCode: "rd-center", Item: model.Department{Name: "后端组", Code: "backend-team", Leader: "陈后端", Phone: "13800000003", Status: "enabled", Sort: 1, Remark: "后端研发"}},
	}
	codeToID := map[string]uint{}
	for _, item := range items {
		var parentID uint
		if item.ParentCode != "" {
			parentID = codeToID[item.ParentCode]
		}
		item.Item.ParentID = parentID
		var department model.Department
		db.Where("code = ?", item.Item.Code).FirstOrCreate(&department, item.Item)
		codeToID[item.Item.Code] = department.ID
	}
}

func seedNotices(db *gorm.DB) {
	items := []model.Notice{
		{Title: "系统升级通知", Content: "本周五晚间进行版本升级，请提前保存数据。", Level: "important", Status: "published", Publisher: "系统管理员", Sort: 1, Remark: "升级公告"},
		{Title: "权限调整提醒", Content: "近期将统一梳理角色权限，请各部门负责人确认菜单访问范围。", Level: "normal", Status: "published", Publisher: "运维中心", Sort: 2, Remark: "权限公告"},
	}
	for _, item := range items {
		db.Where("title = ?", item.Title).FirstOrCreate(&model.Notice{}, item)
	}
}

func ensureDemoUsers(db *gorm.DB, mode string) {
	// 仅在 debug 模式下兜底修复测试账号，避免历史脏数据导致无法登录
	if mode != "debug" {
		return
	}
	defaultHash := "$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy" // 123456
	type seedUser struct {
		Username string
		Role     string
		Nickname string
		Avatar   string
	}
	users := []seedUser{
		{Username: "admin", Role: "admin", Nickname: "系统管理员", Avatar: "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png"},
		{Username: "editor", Role: "editor", Nickname: "内容编辑", Avatar: "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_2.png"},
		{Username: "test", Role: "test", Nickname: "测试账号", Avatar: "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_3.png"},
	}

	var roles []model.Role
	db.Find(&roles)
	roleMap := map[string]uint{}
	for _, r := range roles {
		roleMap[r.Code] = r.ID
	}

	for _, u := range users {
		var exists model.User
		err := db.Where("username = ?", u.Username).First(&exists).Error
		if err == nil {
			db.Model(&exists).Updates(map[string]any{
				"password_hash": defaultHash,
				"role":          u.Role,
				"nickname":      u.Nickname,
				"avatar":        u.Avatar,
			})
		} else {
			newUser := model.User{
				Username:     u.Username,
				PasswordHash: defaultHash,
				Role:         u.Role,
				Nickname:     u.Nickname,
				Avatar:       u.Avatar,
			}
			db.Create(&newUser)
			exists = newUser
		}
		if rid, ok := roleMap[u.Role]; ok && exists.ID > 0 {
			db.FirstOrCreate(&model.UserRole{}, model.UserRole{UserID: exists.ID, RoleID: rid})
		}
	}
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf("redis addr is empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

func initNSQ(cfg *config.Config) (*nsq.Producer, error) {
	if cfg.NSQ.ProducerAddr == "" {
		return nil, fmt.Errorf("nsq producer addr is empty")
	}
	producer, err := nsq.NewProducer(cfg.NSQ.ProducerAddr, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func initStorage(cfg *config.Config) (storage.Client, error) {
	switch cfg.Storage.Mode {
	case "aws-s3":
		return storage.NewMinIOS3(storage.MinIOS3Config{
			Endpoint:      cfg.Storage.S3.Endpoint,
			AccessKey:     cfg.Storage.S3.AccessKey,
			SecretKey:     cfg.Storage.S3.SecretKey,
			UseSSL:        cfg.Storage.S3.UseSSL,
			PublicBucket:  cfg.Storage.S3.PublicBucket,
			PrivateBucket: cfg.Storage.S3.PrivateBucket,
		})
	case "minio":
		return storage.NewMinIOS3(storage.MinIOS3Config{
			Endpoint:      cfg.Storage.MinIO.Endpoint,
			AccessKey:     cfg.Storage.MinIO.AccessKey,
			SecretKey:     cfg.Storage.MinIO.SecretKey,
			UseSSL:        cfg.Storage.MinIO.UseSSL,
			PublicBucket:  cfg.Storage.MinIO.PublicBucket,
			PrivateBucket: cfg.Storage.MinIO.PrivateBucket,
		})
	default:
		return storage.NewLocal(cfg.Storage.Local.BaseDir), nil
	}
}
