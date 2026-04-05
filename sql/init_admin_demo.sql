-- admin_demo baseline schema for internal admin systems

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(32) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  phone VARCHAR(20) DEFAULT '',
  email VARCHAR(128) DEFAULT '',
  nickname VARCHAR(64) DEFAULT '',
  avatar VARCHAR(255) DEFAULT '',
  bio VARCHAR(255) DEFAULT '',
  role VARCHAR(16) NOT NULL DEFAULT 'editor',
  google_secret VARCHAR(64) DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL UNIQUE,
  name VARCHAR(64) NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(16) NOT NULL DEFAULT 'api',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS menus (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  parent_id BIGINT NOT NULL DEFAULT 0,
  path VARCHAR(128) NOT NULL,
  name VARCHAR(64) DEFAULT '',
  component VARCHAR(255) DEFAULT '',
  redirect VARCHAR(128) DEFAULT '',
  title VARCHAR(64) NOT NULL,
  icon VARCHAR(64) DEFAULT '',
  badge VARCHAR(32) DEFAULT '',
  permission_code VARCHAR(64) DEFAULT '',
  always_show TINYINT(1) NOT NULL DEFAULT 0,
  affix TINYINT(1) NOT NULL DEFAULT 0,
  no_keep_alive TINYINT(1) NOT NULL DEFAULT 0,
  hidden TINYINT(1) NOT NULL DEFAULT 0,
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dict_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  dict_type VARCHAR(64) NOT NULL,
  label VARCHAR(128) NOT NULL,
  value VARCHAR(128) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'enabled',
  sort INT NOT NULL DEFAULT 0,
  remark VARCHAR(255) DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_dict_type_value (dict_type, value),
  INDEX idx_dict_type (dict_type)
);

CREATE TABLE IF NOT EXISTS system_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  config_key VARCHAR(64) NOT NULL UNIQUE,
  config_value TEXT NOT NULL,
  name VARCHAR(128) NOT NULL,
  `group` VARCHAR(64) NOT NULL,
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  remark VARCHAR(255) DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_config_group (`group`)
);

CREATE TABLE IF NOT EXISTS operation_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL,
  operator VARCHAR(64) DEFAULT '',
  target VARCHAR(128) DEFAULT '',
  request_id VARCHAR(64) DEFAULT '',
  ip VARCHAR(64) DEFAULT '',
  detail VARCHAR(255) DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_operation_module (module),
  INDEX idx_operation_operator (operator),
  INDEX idx_operation_request_id (request_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id BIGINT NOT NULL,
  permission_id BIGINT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS role_menus (
  role_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, menu_id)
);

-- default roles
INSERT IGNORE INTO roles(code, name) VALUES
('admin', '管理员'),
('editor', '编辑员'),
('test', '测试员');

-- default user password: 123456
INSERT IGNORE INTO users(username, password_hash, role, nickname, avatar) VALUES
('admin', '$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy', 'admin', '系统管理员', 'https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png'),
('editor', '$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy', 'editor', '内容编辑', 'https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_2.png'),
('test', '$2a$10$GQb/mnh9s1mA.bJFpAYqV.M5th1fHhEiSYwU8lvk9c1MlgN6mBUcy', 'test', '测试账号', 'https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_3.png');

INSERT IGNORE INTO dict_items(dict_type, label, value, status, sort, remark) VALUES
('user_status', '启用', 'enabled', 'enabled', 1, '用户状态'),
('user_status', '禁用', 'disabled', 'enabled', 2, '用户状态'),
('notice_level', '普通', 'normal', 'enabled', 1, '通知等级'),
('notice_level', '重要', 'important', 'enabled', 2, '通知等级'),
('storage_mode', '本地存储', 'local', 'enabled', 1, '存储模式'),
('storage_mode', 'MinIO', 'minio', 'enabled', 2, '存储模式');

INSERT IGNORE INTO system_configs(config_key, config_value, name, `group`, value_type, remark) VALUES
('site.title', 'Admin Demo', '站点标题', 'site', 'string', '后台系统标题'),
('site.logo', '/logo.png', '站点 Logo', 'site', 'string', '站点 logo 地址'),
('security.login_captcha', 'false', '登录验证码', 'security', 'boolean', '是否启用登录验证码'),
('storage.default_mode', 'local', '默认存储模式', 'storage', 'string', '默认文件存储模式');
