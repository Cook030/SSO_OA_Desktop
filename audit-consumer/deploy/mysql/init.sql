-- 本地联调用: 为 Canal 创建最小权限账号(与 instance.properties 账号一致)
-- 生产环境请在 RDS/数据库控制台执行并修改为强密码
CREATE USER IF NOT EXISTS 'canal'@'%' IDENTIFIED BY 'canal123';
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';
FLUSH PRIVILEGES;
