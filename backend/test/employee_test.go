package test

import (
	"fmt"
	"testing"

	"permission-system/test/testutil"
)

// TestEmployeeBusiness 按业务组织员工管理全生命周期测试
func TestEmployeeBusiness(t *testing.T) {
	// 1. 创建员工
	employeeID, account, cleanup := testutil.CreateEmployee()
	defer cleanup()

	// 2. 查询员工列表，验证创建成功
	resp := testutil.E.GET("/employees").WithQuery("pageSize", 100).Expect()
	obj := testutil.MustOK(resp)
	list := obj.Value("data").Object().Value("list").Array()
	found := false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("员工列表中未找到刚创建的员工: %d", employeeID)
	}

	// 3. 部门列表应包含测试部
	resp = testutil.E.GET("/employees/departments").Expect()
	testutil.MustOK(resp)
	depts := resp.JSON().Object().Value("data").Array()
	deptFound := false
	for i := 0; i < int(depts.Length().Raw()); i++ {
		if depts.Value(i).String().Raw() == "测试部" {
			deptFound = true
			break
		}
	}
	if !deptFound {
		t.Fatalf("部门列表中未找到测试部")
	}

	// 4. 修改员工
	newName := fmt.Sprintf("员工%d", 1000000+1)
	newPhone := testutil.RandomPhone()
	newEmailPrefix := testutil.RandomEmailPrefix()
	newAccount := testutil.RandomAccount()
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", employeeID).WithJSON(map[string]any{
		"name":        newName,
		"phone":       newPhone,
		"emailPrefix": newEmailPrefix,
		"account":     newAccount,
		"department":  "测试部",
		"roleIds":     []int64{},
	}).Expect()
	obj = testutil.MustOK(resp)
	data := obj.Value("data").Object()
	data.Value("id").Number().IsEqual(float64(employeeID))
	data.Value("name").String().IsEqual(newName)
	data.Value("phone").String().IsEqual(newPhone)
	data.Value("email").String().IsEqual(newEmailPrefix + testutil.EmailDomain)
	account = newAccount

	// 5. 再次查询，验证修改生效
	resp = testutil.E.GET("/employees").WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	found = false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			item.Value("name").String().IsEqual(newName)
			item.Value("phone").String().IsEqual(newPhone)
			item.Value("department").String().IsEqual("测试部")
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("修改后的员工未在列表中找到: %d", employeeID)
	}

	// 6. 重置员工密码（会同步调用 SSO 撤销该用户会话）
	resp = testutil.E.PUT("/employees/{id}/reset-password").WithPath("id", employeeID).Expect()
	testutil.MustOK(resp)

	// 7. 参数错误：员工姓名过短
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "A",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "测试部",
		"password":    "Test@123456",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 8. 参数错误：手机号格式不正确
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "张三",
		"phone":       "12345678901",
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "测试部",
		"password":    "Test@123456",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 9. 参数错误：所属部门为空
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "张三",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"department":  "",
		"password":    "Test@123456",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 10. 参数错误：账号为空
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "张三",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     "",
		"department":  "测试部",
		"password":    "Test@123456",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 11. 参数错误：密码过短
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "张三",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "测试部",
		"password":    "123",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 12. 重复数据：使用相同账号创建
	resp = testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        "李四",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     account,
		"department":  "测试部",
		"password":    "Test@123456",
	}).Expect()
	obj = testutil.MustCode(resp, 400)
	testutil.MustMessageContains(obj, "该邮箱前缀(账号)已存在")

	// 13. 重复数据：更新时手机号冲突
	otherID, _, otherCleanup := testutil.CreateEmployee()
	defer otherCleanup()
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", otherID).WithJSON(map[string]any{
		"name":        "李四",
		"phone":       newPhone,
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "测试部",
		"roleIds":     []int64{},
	}).Expect()
	testutil.MustCode(resp, 400)

	// 14. 资源不存在：修改不存在的员工（实际返回 400，message 含"不存在"）
	resp = testutil.E.PUT("/employees/{id}").WithPath("id", 9999999).WithJSON(map[string]any{
		"name":        "不存在",
		"phone":       testutil.RandomPhone(),
		"emailPrefix": testutil.RandomEmailPrefix(),
		"account":     testutil.RandomAccount(),
		"department":  "测试部",
		"roleIds":     []int64{},
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 15. 资源不存在：删除不存在的员工（实际返回 400，message 含"不存在"）
	resp = testutil.E.DELETE("/employees/{id}").WithPath("id", 9999999).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 16. 未登录
	testutil.MustCode(testutil.NoAuthE.GET("/employees").Expect(), 401)
	testutil.MustCode(testutil.NoAuthRequest("POST", "/employees", map[string]any{
		"name": "张三", "phone": testutil.RandomPhone(), "emailPrefix": testutil.RandomEmailPrefix(), "account": testutil.RandomAccount(), "department": "测试部",
	}), 401)
	testutil.MustCode(testutil.NoAuthRequest("PUT", fmt.Sprintf("/employees/%d", employeeID), map[string]any{
		"name": "张三", "phone": testutil.RandomPhone(), "emailPrefix": testutil.RandomEmailPrefix(), "account": testutil.RandomAccount(), "department": "测试部",
	}), 401)
	testutil.MustCode(testutil.NoAuthRequest("DELETE", fmt.Sprintf("/employees/%d", employeeID)), 401)

	// 17. Token 无效
	testutil.MustCode(testutil.NoAuthE.GET("/employees").WithCookie(testutil.CookieAccessTokenName, testutil.TokenInvalidValue).Expect(), 401)
	testutil.MustCode(testutil.NoAuthE.GET("/employees/departments").WithCookie(testutil.CookieAccessTokenName, testutil.TokenInvalidValue).Expect(), 401)

	// 18. 权限不足：员工访问员工管理接口
	testutil.MustCode(testutil.EmployeeRequest("GET", "/employees"), 403)
	testutil.MustCode(testutil.EmployeeRequest("GET", "/employees/departments"), 403)
	testutil.MustCode(testutil.EmployeeRequest("POST", "/employees", map[string]any{
		"name": "张三", "phone": testutil.RandomPhone(), "emailPrefix": testutil.RandomEmailPrefix(), "account": testutil.RandomAccount(), "department": "测试部",
	}), 403)
	testutil.MustCode(testutil.EmployeeRequest("PUT", fmt.Sprintf("/employees/%d", employeeID), map[string]any{
		"name": "张三", "phone": testutil.RandomPhone(), "emailPrefix": testutil.RandomEmailPrefix(), "account": testutil.RandomAccount(), "department": "测试部",
	}), 403)
	testutil.MustCode(testutil.EmployeeRequest("DELETE", fmt.Sprintf("/employees/%d", employeeID)), 403)

	// 19. 删除员工
	testutil.DeleteEmployee(employeeID)

	// 20. 再次查询，验证删除生效
	resp = testutil.E.GET("/employees").WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == employeeID {
			t.Fatalf("已删除的员工仍出现在列表中: %d", employeeID)
		}
	}

	// 21. 再次删除，验证资源不存在（实际返回 400，message 含"不存在"）
	resp = testutil.E.DELETE("/employees/{id}").WithPath("id", employeeID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")
}
