package test

import (
	"testing"

	"permission-system/test/testutil"
)

// TestFullFlow 完整业务流程测试（RBAC 版本）
func TestFullFlow(t *testing.T) {
	// 1. 管理员登录（已在 TestMain 中完成，这里复用 E 客户端）
	// 2. 创建平台
	platformID, platformCleanup := testutil.CreatePlatform()
	defer platformCleanup()

	// 3. 创建员工
	employeeID, _, employeeCleanup := testutil.CreateEmployee()
	defer employeeCleanup()

	// 4. 查询员工列表，验证创建成功
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
		"roleIds":     []int64{},
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

	// 7. 创建测试角色
	roleCode := "test_flow_role"
	roleResp := testutil.E.POST("/roles").WithJSON(map[string]any{
		"code":        roleCode,
		"name":        "流程测试角色",
		"description": "由流程测试创建",
	}).Expect()
	testutil.MustOK(roleResp)
	roleID := int64(roleResp.JSON().Object().Value("data").Object().Value("id").Number().Raw())
	defer func() {
		testutil.E.DELETE("/roles/{id}").WithPath("id", roleID).Expect().Status(200)
	}()

	// 8. 给员工分配该角色
	assignResp := testutil.E.POST("/roles/users").WithJSON(map[string]any{
		"userIds": []int64{employeeID},
		"roleIds": []int64{roleID},
	}).Expect()
	testutil.MustOK(assignResp)

	// 9. 管理员 /me/permissions 应返回 admin 角色与完整权限
	meResp := testutil.E.GET("/me/permissions").Expect()
	meObj := testutil.MustOK(meResp).Value("data").Object()
	meObj.Value("isAdmin").Boolean().IsEqual(true)
	roles := meObj.Value("roles").Array()
	hasAdmin := false
	for i := 0; i < int(roles.Length().Raw()); i++ {
		if roles.Value(i).String().Raw() == "admin" {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		t.Fatalf("流程测试：管理员 /me/permissions 未返回 admin 角色")
	}

	// 10. 删除员工
	testutil.DeleteEmployee(employeeID)

	// 11. 验证资源不存在
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", employeeID).WithJSON(map[string]any{
		"name":        newName,
		"phone":       newPhone,
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "流程测试部",
		"roleIds":     []int64{},
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	resp = testutil.E.DELETE("/employees/{id}").WithPath("id", employeeID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 12. 删除平台
	testutil.DeletePlatform(platformID)

	// 13. 验证资源不存在
	resp = testutil.E.PUT("/platforms/{id}").WithPath("id", platformID).WithJSON(map[string]any{
		"name": testutil.RandomPlatformName(),
		"code": testutil.RandomPlatformCode(),
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	resp = testutil.E.DELETE("/platforms/{id}").WithPath("id", platformID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")
}
