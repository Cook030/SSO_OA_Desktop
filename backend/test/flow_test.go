package test

import (
	"testing"

	"permission-system/test/testutil"
)

// TestFullFlow 完整业务流程测试
func TestFullFlow(t *testing.T) {
	// 1. 管理员登录（已在 TestMain 中完成，这里复用 E 客户端）
	// 2. 创建平台
	platformID, platformCleanup := testutil.CreatePlatform()
	defer platformCleanup()

	// 3. 创建员工
	employeeID, _, employeeCleanup := testutil.CreateEmployee()
	defer employeeCleanup()

	// 4. 查询员工
	resp := testutil.E.GET("/employees").WithQuery("pageSize", 100).Expect()
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
		t.Fatalf("流程测试：员工列表中未找到刚创建的员工: %d", employeeID)
	}

	// 5. 修改员工
	newName := "流程测试员工-更新"
	newPhone := testutil.RandomPhone()
	newEmailPrefix := testutil.RandomEmailPrefix()
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", employeeID).WithJSON(map[string]any{
		"name":        newName,
		"phone":       newPhone,
		"emailPrefix": newEmailPrefix,
		"department":  "流程测试部",
		"platformIds": []int64{},
	}).Expect()
	testutil.MustOK(resp)

	// 6. 再次查询
	resp = testutil.E.GET("/employees").WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	found = false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			item.Value("name").String().IsEqual(newName)
			item.Value("phone").String().IsEqual(newPhone)
			item.Value("department").String().IsEqual("流程测试部")
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("流程测试：修改后的员工未在列表中找到: %d", employeeID)
	}

	// 7. 批量新增权限
	affected := testutil.GrantPermission([]int64{employeeID}, []int64{platformID})
	if affected <= 0 {
		t.Fatalf("流程测试：批量新增权限应至少影响 1 条记录，实际: %d", affected)
	}

	// 8. 验证权限：按平台筛选员工列表应包含该员工
	resp = testutil.E.GET("/employees").WithQuery("platformId", platformID).WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	found = false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("流程测试：按平台筛选未找到已授权员工: %d", employeeID)
	}

	// 9. 批量删除权限
	affected = testutil.RemovePermission([]int64{employeeID}, []int64{platformID})
	if affected <= 0 {
		t.Fatalf("流程测试：批量删除权限应至少影响 1 条记录，实际: %d", affected)
	}

	// 10. 验证权限删除：按平台筛选员工列表应不再包含该员工
	resp = testutil.E.GET("/employees").WithQuery("platformId", platformID).WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			t.Fatalf("流程测试：按平台筛选仍包含已移除权限的员工: %d", employeeID)
		}
	}

	// 11. 删除员工
	testutil.DeleteEmployee(employeeID)

	// 12. 验证资源不存在（实际返回 400，message 含"不存在"）
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", employeeID).WithJSON(map[string]any{
		"name":        newName,
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "流程测试部",
		"platformIds": []int64{},
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	resp = testutil.E.DELETE("/employees/{id}").WithPath("id", employeeID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 13. 删除平台
	testutil.DeletePlatform(platformID)

	// 14. 验证资源不存在（实际返回 400，message 含"不存在"）
	resp = testutil.E.PUT("/platforms/{id}").WithPath("id", platformID).WithJSON(map[string]any{
		"name": testutil.RandomPlatformName(),
		"link": testutil.RandomPlatformLink(),
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	resp = testutil.E.DELETE("/platforms/{id}").WithPath("id", platformID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")
}
