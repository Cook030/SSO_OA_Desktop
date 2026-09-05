package shared

// EmailDomain 员工邮箱域名后缀
const EmailDomain = "@maplehaze.cn"

// DefaultPassword 新员工默认密码
const DefaultPassword = "Mhint@123"

// ---------- 通用状态 ----------

// 启用/禁用状态（sys_role.status、sys_permission.status 共用）
const (
	StatusEnabled  int32 = 1
	StatusDisabled int32 = 0
)

// ---------- 权限类型 ----------

// sys_permission.type
const (
	PermissionTypeMenu int32 = 1 // 菜单权限（前端展示用，不参与接口鉴权）
	PermissionTypeAPI  int32 = 2 // API 权限（参与接口鉴权）
)

// ---------- 内置角色 ----------

// 内置角色编码，与 sys_role.code 对齐；内置角色（is_builtin=1）不可删除
const (
	RoleAdmin    = "admin"    // 超级管理员：拥有全部权限
	RoleHR       = "hr"       // 人事：员工与平台管理
	RoleManager  = "manager"  // 部门主管：只读
	RoleEmployee = "employee" // 普通员工：无后台权限
)

// BuiltinRoles 内置角色定义，用于启动时的角色初始化
var BuiltinRoles = []struct {
	Code        string
	Name        string
	Description string
}{
	{RoleAdmin, "超级管理员", "拥有全部权限，不可删除"},
	{RoleHR, "人事", "可管理平台与员工"},
	{RoleManager, "部门主管", "可查看平台与员工信息"},
	{RoleEmployee, "普通员工", "默认角色，无后台管理权限"},
}

// ---------- 权限码 ----------

// Wildcard 通配符。支持 "*:*"(全部权限) 与 "<object>:*"(某资源的全部动作)
const Wildcard = "*"

// 平台访问权限码格式：platform:<平台编码>:access
// 平台自身的访问权限同样是一枚权限点，通过角色授予，从而被 RBAC 统一承载。
const (
	PlatformAccessPrefix = "platform:"
	PlatformAccessSuffix = ":access"
)

// PlatformAccessCode 生成平台访问权限码
func PlatformAccessCode(platformCode string) string {
	return PlatformAccessPrefix + platformCode + PlatformAccessSuffix
}
