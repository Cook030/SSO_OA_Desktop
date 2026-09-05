-- 用户表
CREATE TABLE IF NOT EXISTS sys_user (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  account     VARCHAR(64)  NOT NULL COMMENT '登录账号',
  password    VARCHAR(255) NOT NULL COMMENT '密码(bcrypt加密)',
  name        VARCHAR(64)  NOT NULL COMMENT '姓名',
  phone       VARCHAR(20)  NOT NULL COMMENT '手机号',
  email       VARCHAR(128) NOT NULL COMMENT '邮箱',
  role             VARCHAR(32)  NOT NULL DEFAULT 'employee' COMMENT '角色: admin/employee',
  department       VARCHAR(64)  DEFAULT NULL COMMENT '所属部门',
  password_changed BOOL         NOT NULL DEFAULT 0 COMMENT '是否已修改初始密码',
  create_time      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY uk_account (account),
  UNIQUE KEY uk_phone (phone),
  UNIQUE KEY uk_email (email),
  KEY idx_department (department)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表(管理员+员工)';


-- 平台表
CREATE TABLE IF NOT EXISTS sys_platform (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(128) NOT NULL COMMENT '平台名称',
  link        VARCHAR(128) NOT NULL COMMENT '平台链接',
  create_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY uk_name (name),
  UNIQUE KEY uk_link (link)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平台表';


-- 用户平台权限关系表
CREATE TABLE IF NOT EXISTS sys_user_platform (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  platform_id BIGINT UNSIGNED NOT NULL COMMENT '平台ID',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE KEY uk_user_platform (user_id, platform_id),
  KEY idx_user_id (user_id),
  KEY idx_platform_id (platform_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户平台权限关系表';
