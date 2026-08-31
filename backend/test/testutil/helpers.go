package testutil

import (
	"fmt"
	"math/rand"
	"time"
)

// ─── 认证与用户信息 ───────────────────────────────────────────────

// MockSSOTokenPrefix mock SSO 服务识别的 token 前缀
const MockSSOTokenPrefix = "mock-sso-token-"

// SSOTokenFor 生成指定账号的 mock SSO Token
// mock SSO 服务根据前缀后的账号查出本地用户 id 返回（与 sys_user.id 对齐）
func SSOTokenFor(account string) string {
	return MockSSOTokenPrefix + account
}

// ─── 随机数据生成 ─────────────────────────────────────────────────

// RandomPhone 生成随机中国大陆手机号
func RandomPhone() string {
	prefixes := []string{"138", "139", "150", "151", "152", "157", "158", "159", "182", "183", "187", "188"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := fmt.Sprintf("%08d", rand.Intn(100000000))
	return prefix + suffix
}

// RandomEmailPrefix 生成随机邮箱前缀
func RandomEmailPrefix() string {
	return fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// RandomAccount 生成随机员工账号(姓名全拼)
func RandomAccount() string {
	return fmt.Sprintf("account_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// RandomEmail 生成随机完整邮箱
func RandomEmail() string {
	return RandomEmailPrefix() + EmailDomain
}

// RandomPlatformName 生成随机平台名称
func RandomPlatformName() string {
	return fmt.Sprintf("平台_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// RandomPlatformLink 生成随机平台链接
func RandomPlatformLink() string {
	return fmt.Sprintf("link_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// RandomEmployee 生成随机员工姓名、手机号、邮箱前缀、账号
func RandomEmployee() (name, phone, emailPrefix, account string) {
	return fmt.Sprintf("员工%d", rand.Intn(1000000)), RandomPhone(), RandomEmailPrefix(), RandomAccount()
}

// ─── 平台 CRUD ────────────────────────────────────────────────────

// CreatePlatform 创建平台并返回 ID 和清理函数
func CreatePlatform() (int64, func()) {
	name := RandomPlatformName()
	link := RandomPlatformLink()

	resp := E.POST("/platforms").WithJSON(map[string]any{
		"name": name,
		"link": link,
	}).Expect()

	data := MustOK(resp).Value("data").Object()
	id := int64(data.Value("id").Number().Raw())

	return id, func() { DeletePlatform(id) }
}

// DeletePlatform 删除指定平台
func DeletePlatform(id int64) {
	if id <= 0 {
		return
	}
	E.DELETE("/platforms/{id}").WithPath("id", id).Expect().Status(200)
}

// ─── 员工 CRUD ────────────────────────────────────────────────────

// CreateEmployee 创建员工并返回 ID、账号和清理函数
func CreateEmployee() (int64, string, func()) {
	name, phone, emailPrefix, account := RandomEmployee()

	resp := E.POST("/employees").WithJSON(map[string]any{
		"name":        name,
		"phone":       phone,
		"emailPrefix": emailPrefix,
		"account":     account,
		"department":  "测试部",
		"platformIds": []int64{},
		"password":    "Test@123456",
	}).Expect()

	data := MustOK(resp).Value("data").Object()
	id := int64(data.Value("id").Number().Raw())
	returnedAccount := data.Value("account").String().Raw()

	return id, returnedAccount, func() { DeleteEmployee(id) }
}

// DeleteEmployee 删除指定员工
func DeleteEmployee(id int64) {
	if id <= 0 {
		return
	}
	E.DELETE("/employees/{id}").WithPath("id", id).Expect().Status(200)
}

// ─── 权限操作 ─────────────────────────────────────────────────────

// GrantPermission 批量为员工授予平台权限
func GrantPermission(userIDs, platformIDs []int64) int64 {
	resp := E.POST("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     userIDs,
		"platformIds": platformIDs,
	}).Expect()

	data := MustOK(resp).Value("data").Object()
	return int64(data.Value("affectedCount").Number().Raw())
}

// RemovePermission 批量删除员工的平台权限
func RemovePermission(userIDs, platformIDs []int64) int64 {
	resp := E.DELETE("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     userIDs,
		"platformIds": platformIDs,
	}).Expect()

	data := MustOK(resp).Value("data").Object()
	return int64(data.Value("affectedCount").Number().Raw())
}
