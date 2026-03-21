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
