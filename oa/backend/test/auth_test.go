package test

import (
	"testing"

	"permission-system/test/testutil"
)

// TestAdminPermission 测试管理员权限拦截
func TestAdminPermission(t *testing.T) {
	// 员工访问管理员接口 → 403
	resp := testutil.EmployeeE.GET("/platforms").Expect()
	obj := testutil.MustCode(resp, 403)
	testutil.MustMessageContains(obj, "无权限")

	// 管理员访问管理员接口 → 200
	resp = testutil.E.GET("/platforms").Expect()
	testutil.MustOK(resp)
}
