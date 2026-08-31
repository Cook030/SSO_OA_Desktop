package test

import (
	"testing"

	"permission-system/test/testutil"
)

// TestPermissionBusiness 按业务组织权限管理全生命周期测试
func TestPermissionBusiness(t *testing.T) {
	// 1. 创建平台
	platformID, platformCleanup := testutil.CreatePlatform()
	defer platformCleanup()

	// 2. 创建员工
	employeeID, _, employeeCleanup := testutil.CreateEmployee()
	defer employeeCleanup()

	// 3. 赋予权限
	affected := testutil.GrantPermission([]int64{employeeID}, []int64{platformID})
	if affected <= 0 {
		t.Fatalf("批量设置权限应至少影响 1 条记录，实际: %d", affected)
	}

	// 4. 员工列表按平台筛选应包含该员工
	resp := testutil.E.GET("/employees").WithQuery("platformId", platformID).WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list := resp.JSON().Object().Value("data").Object().Value("list").Array()
	found := false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("按平台筛选员工列表未找到已授权员工: %d", employeeID)
	}

	// 5. 删除权限
	affected = testutil.RemovePermission([]int64{employeeID}, []int64{platformID})
	if affected <= 0 {
		t.Fatalf("批量删除权限应至少影响 1 条记录，实际: %d", affected)
	}

	// 6. 按平台筛选员工列表应不再包含该员工
	resp = testutil.E.GET("/employees").WithQuery("platformId", platformID).WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			t.Fatalf("按平台筛选仍包含已移除权限的员工: %d", employeeID)
		}
	}

	// 7. 参数错误：userIds 为空
	resp = testutil.E.POST("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     []int64{},
		"platformIds": []int64{platformID},
	}).Expect()
	testutil.MustCode(resp, 400)

	// 8. 参数错误：platformIds 为空
	resp = testutil.E.POST("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{},
	}).Expect()
	testutil.MustCode(resp, 400)

	// 9. 参数错误：批量删除时 userIds 为空
	resp = testutil.E.DELETE("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     []int64{},
		"platformIds": []int64{platformID},
	}).Expect()
	testutil.MustCode(resp, 400)

	// 10. 参数错误：批量删除时 platformIds 为空
	resp = testutil.E.DELETE("/employees/permissions/batch").WithJSON(map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{},
	}).Expect()
	testutil.MustCode(resp, 400)

	// 11. 未登录
	testutil.MustCode(testutil.NoAuthRequest("POST", "/employees/permissions/batch", map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{platformID},
	}), 401)
	testutil.MustCode(testutil.NoAuthRequest("DELETE", "/employees/permissions/batch", map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{platformID},
	}), 401)

	// 12. Token 无效
	resp = testutil.NoAuthE.POST("/employees/permissions/batch").WithCookie(testutil.CookieAccessTokenName, testutil.TokenInvalidValue).WithJSON(map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{platformID},
	}).Expect()
	testutil.MustCode(resp, 401)

	// 13. 权限不足：员工访问权限管理接口
	testutil.MustCode(testutil.EmployeeRequest("POST", "/employees/permissions/batch", map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{platformID},
	}), 403)
	testutil.MustCode(testutil.EmployeeRequest("DELETE", "/employees/permissions/batch", map[string]any{
		"userIds":     []int64{employeeID},
		"platformIds": []int64{platformID},
	}), 403)
}
