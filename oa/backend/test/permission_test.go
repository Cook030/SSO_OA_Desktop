package test

import (
	"encoding/json"
	"testing"

	"permission-system/test/testutil"
)

// permissionTestNode 权限树节点（仅供测试解析使用）
type permissionTestNode struct {
	ID       int64                 `json:"id"`
	Code     string                `json:"code"`
	Children []permissionTestNode  `json:"children"`
}

// findPermissionIDByCode 递归查找权限点 ID
func findPermissionIDByCode(tree []permissionTestNode, code string) int64 {
	for _, n := range tree {
		if n.Code == code {
			return n.ID
		}
		if id := findPermissionIDByCode(n.Children, code); id != 0 {
			return id
		}
	}
	return 0
}

// TestPermissionBusiness 角色与权限管理全生命周期测试（RBAC 版本）
func TestPermissionBusiness(t *testing.T) {
	// 1. 创建测试平台
	platformID, platformCleanup := testutil.CreatePlatform()
	defer platformCleanup()

	// 2. 创建测试员工
	employeeID, _, employeeCleanup := testutil.CreateEmployee()
	defer employeeCleanup()

	// 3. 创建测试角色
	roleCode := "test_perm_role"
	roleResp := testutil.E.POST("/roles").WithJSON(map[string]any{
		"code":        roleCode,
		"name":        "权限测试角色",
		"description": "由权限测试创建",
	}).Expect()
	obj := testutil.MustOK(roleResp)
	roleID := int64(obj.Value("data").Object().Value("id").Number().Raw())
	defer func() {
		testutil.E.DELETE("/roles/{id}").WithPath("id", roleID).Expect().Status(200)
	}()

	// 4. 查询角色列表，应包含新角色
	listResp := testutil.E.GET("/roles").Expect()
	obj = testutil.MustOK(listResp)
	roleList := obj.Value("data").Array()
	found := false
	for i := 0; i < int(roleList.Length().Raw()); i++ {
		item := roleList.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == roleID {
			found = true
			item.Value("code").String().IsEqual(roleCode)
			break
		}
	}
	if !found {
		t.Fatalf("角色列表中未找到刚创建的角色: %d", roleID)
	}

	// 5. 从权限树中查找 role:list 的 ID
	permsResp := testutil.E.GET("/permissions").Expect()
	testutil.MustOK(permsResp)
	var permTree []permissionTestNode
	if err := json.Unmarshal([]byte(permsResp.Body().Raw()), &permTree); err != nil {
		t.Fatalf("解析权限树失败: %v", err)
	}
	roleListPermID := findPermissionIDByCode(permTree, "role:list")
	if roleListPermID == 0 {
		t.Fatalf("未在权限树中找到 code=role:list")
	}

	// 6. 给角色配置一个权限
	assignResp := testutil.E.PUT("/roles/{id}/permissions").WithPath("id", roleID).WithJSON(map[string]any{
		"permissionIds": []int64{roleListPermID},
	}).Expect()
	testutil.MustOK(assignResp)

	// 7. 查询角色已分配的权限
	permsOfRoleResp := testutil.E.GET("/roles/{id}/permissions").WithPath("id", roleID).Expect()
	obj = testutil.MustOK(permsOfRoleResp)
	permsOfRole := obj.Value("data").Array()
	if int(permsOfRole.Length().Raw()) != 1 {
		t.Fatalf("角色分配权限后应仅有 1 项，实际: %d", int(permsOfRole.Length().Raw()))
	}
	if int64(permsOfRole.Value(0).Number().Raw()) != roleListPermID {
		t.Fatalf("角色分配权限 ID 不匹配")
	}

	// 8. 给员工分配该角色
	assignUserResp := testutil.E.POST("/roles/users").WithJSON(map[string]any{
		"userIds": []int64{employeeID},
		"roleIds": []int64{roleID},
	}).Expect()
	obj = testutil.MustOK(assignUserResp)
	if int64(obj.Value("data").Object().Value("affectedCount").Number().Raw()) < 1 {
		t.Fatalf("批量给用户分配角色应至少影响 1 条记录")
	}

	// 9. 查询员工当前角色
	userRolesResp := testutil.E.GET("/users/{id}/roles").WithPath("id", employeeID).Expect()
	obj = testutil.MustOK(userRolesResp)
	userRoles := obj.Value("data").Array()
	found = false
	for i := 0; i < int(userRoles.Length().Raw()); i++ {
		if int64(userRoles.Value(i).Object().Value("id").Number().Raw()) == roleID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("员工未持有刚分配的角色: %d", roleID)
	}

	// 10. 全量覆盖设置员工角色
	setResp := testutil.E.PUT("/users/{id}/roles").WithPath("id", employeeID).WithJSON(map[string]any{
		"roleIds": []int64{roleID},
	}).Expect()
	testutil.MustOK(setResp)

	// 11. 参数错误：分配角色时 userIds 为空
	testutil.MustCode(testutil.E.POST("/roles/users").WithJSON(map[string]any{
		"userIds": []int64{},
		"roleIds": []int64{roleID},
	}).Expect(), 400)

	// 12. 参数错误：分配角色时 roleIds 为空
	testutil.MustCode(testutil.E.POST("/roles/users").WithJSON(map[string]any{
		"userIds": []int64{employeeID},
		"roleIds": []int64{},
	}).Expect(), 400)

	// 13. 未登录
	testutil.MustCode(testutil.NoAuthRequest("GET", "/roles", nil), 401)
	testutil.MustCode(testutil.NoAuthRequest("POST", "/roles", map[string]any{
		"code": "noauth", "name": "noauth", "description": "",
	}), 401)
	testutil.MustCode(testutil.NoAuthRequest("POST", "/roles/users", map[string]any{
		"userIds": []int64{employeeID}, "roleIds": []int64{roleID},
	}), 401)

	// 14. Token 无效
	testutil.MustCode(testutil.NoAuthE.GET("/roles").WithCookie(testutil.CookieAccessTokenName, testutil.TokenInvalidValue).Expect(), 401)

	// 15. 权限不足：员工访问角色管理接口
	testutil.MustCode(testutil.EmployeeRequest("GET", "/roles"), 403)
	testutil.MustCode(testutil.EmployeeRequest("POST", "/roles", map[string]any{
		"code": "denied", "name": "denied", "description": "",
	}), 403)
	testutil.MustCode(testutil.EmployeeRequest("POST", "/roles/users", map[string]any{
		"userIds": []int64{employeeID}, "roleIds": []int64{roleID},
	}), 403)

	// 16. 内置角色不可删除
	builtinListResp := testutil.E.GET("/roles").Expect()
	obj = testutil.MustOK(builtinListResp)
	roles := obj.Value("data").Array()
	var adminID int64
	for i := 0; i < int(roles.Length().Raw()); i++ {
		item := roles.Value(i).Object()
		if item.Value("code").String().Raw() == "admin" {
			adminID = int64(item.Value("id").Number().Raw())
			break
		}
	}
	if adminID > 0 {
		del := testutil.E.DELETE("/roles/{id}").WithPath("id", adminID).Expect()
		obj = testutil.MustCode(del, 400)
		testutil.MustMessageContains(obj, "内置角色不可删除")
	}

	_ = platformID
}
