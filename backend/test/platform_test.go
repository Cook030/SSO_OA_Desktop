package test

import (
	"fmt"
	"testing"

	"permission-system/test/testutil"
)

// TestPlatformBusiness 按业务组织平台管理全生命周期测试
func TestPlatformBusiness(t *testing.T) {
	// 1. 创建平台
	platformID, cleanup := testutil.CreatePlatform()
	defer cleanup()

	// 2. 查询平台列表，验证创建成功
	resp := testutil.E.GET("/platforms").WithQuery("pageSize", 100).Expect()
	obj := testutil.MustOK(resp)
	list := obj.Value("data").Object().Value("list").Array()
	list.Length().Gt(0)
	found := false
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == platformID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("平台列表中未找到刚创建的平台: %d", platformID)
	}

	// 3. 修改平台
	newName := testutil.RandomPlatformName()
	newLink := testutil.RandomPlatformLink()
	resp = testutil.E.PUT("/platforms/{id}").WithPath("id", platformID).WithJSON(map[string]any{
		"name": newName,
		"link": newLink,
	}).Expect()
	obj = testutil.MustOK(resp)
	data := obj.Value("data").Object()
	data.Value("id").Number().IsEqual(float64(platformID))
	data.Value("name").String().IsEqual(newName)
	data.Value("link").String().IsEqual(newLink)

	// 4. 再次查询，验证修改生效
	resp = testutil.E.GET("/platforms").WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	found = false
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == platformID {
			item.Value("name").String().IsEqual(newName)
			item.Value("link").String().IsEqual(newLink)
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("修改后的平台未在列表中找到: %d", platformID)
	}

	// 5. 参数错误：创建平台名称过短
	resp = testutil.E.POST("/platforms").WithJSON(map[string]any{
		"name": "A",
		"link": newLink,
	}).Expect()
	testutil.MustCode(resp, 400)

	// 6. 参数错误：创建平台链接过短
	resp = testutil.E.POST("/platforms").WithJSON(map[string]any{
		"name": testutil.RandomPlatformName(),
		"link": "A",
	}).Expect()
	testutil.MustCode(resp, 400)

	// 7. 重复数据：使用相同名称创建
	resp = testutil.E.POST("/platforms").WithJSON(map[string]any{
		"name": newName,
		"link": testutil.RandomPlatformLink(),
	}).Expect()
	obj = testutil.MustCode(resp, 400)
	testutil.MustMessageContains(obj, "平台名称已存在")

	// 8. 重复数据：使用相同链接创建
	resp = testutil.E.POST("/platforms").WithJSON(map[string]any{
		"name": testutil.RandomPlatformName(),
		"link": newLink,
	}).Expect()
	obj = testutil.MustCode(resp, 400)
	testutil.MustMessageContains(obj, "平台链接已存在")

	// 9. 资源不存在：修改不存在的平台（实际返回 400，message 含"不存在"）
	resp = testutil.E.PUT("/platforms/{id}").WithPath("id", 9999999).WithJSON(map[string]any{
		"name": testutil.RandomPlatformName(),
		"link": testutil.RandomPlatformLink(),
	}).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 10. 资源不存在：删除不存在的平台（实际返回 400，message 含"不存在"）
	resp = testutil.E.DELETE("/platforms/{id}").WithPath("id", 9999999).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")

	// 11. 重复数据：更新时名称冲突
	otherID, otherCleanup := testutil.CreatePlatform()
	defer otherCleanup()
	resp = testutil.E.PUT("/platforms/{id}").WithPath("id", otherID).WithJSON(map[string]any{
		"name": newName,
		"link": testutil.RandomPlatformLink(),
	}).Expect()
	testutil.MustCode(resp, 400)

	// 12. 未登录
	testutil.MustCode(testutil.NoAuthE.GET("/platforms").Expect(), 401)
	testutil.MustCode(testutil.NoAuthRequest("POST", "/platforms", map[string]any{"name": testutil.RandomPlatformName(), "link": testutil.RandomPlatformLink()}), 401)
	testutil.MustCode(testutil.NoAuthRequest("PUT", "/platforms/1", map[string]any{"name": testutil.RandomPlatformName(), "link": testutil.RandomPlatformLink()}), 401)
	testutil.MustCode(testutil.NoAuthRequest("DELETE", "/platforms/1"), 401)

	// 13. Token 无效
	testutil.MustCode(testutil.NoAuthE.GET("/platforms").WithCookie(testutil.CookieAccessTokenName, testutil.TokenInvalidValue).Expect(), 401)

	// 14. 权限不足：员工访问平台管理接口
	testutil.MustCode(testutil.EmployeeRequest("GET", "/platforms"), 403)
	testutil.MustCode(testutil.EmployeeRequest("POST", "/platforms", map[string]any{"name": testutil.RandomPlatformName(), "link": testutil.RandomPlatformLink()}), 403)
	testutil.MustCode(testutil.EmployeeRequest("PUT", fmt.Sprintf("/platforms/%d", platformID), map[string]any{"name": testutil.RandomPlatformName(), "link": testutil.RandomPlatformLink()}), 403)
	testutil.MustCode(testutil.EmployeeRequest("DELETE", fmt.Sprintf("/platforms/%d", platformID)), 403)

	// 15. 删除平台
	testutil.DeletePlatform(platformID)

	// 16. 再次查询，验证删除生效（列表中不存在）
	resp = testutil.E.GET("/platforms").WithQuery("pageSize", 100).Expect()
	testutil.MustOK(resp)
	list = resp.JSON().Object().Value("data").Object().Value("list").Array()
	for i := 0; i < int(list.Length().Raw()); i++ {
		item := list.Value(i).Object()
		if int64(item.Value("id").Number().Raw()) == platformID {
			t.Fatalf("已删除的平台仍出现在列表中: %d", platformID)
		}
	}

	// 17. 再次删除，验证资源不存在（实际返回 400，message 含"不存在"）
	resp = testutil.E.DELETE("/platforms/{id}").WithPath("id", platformID).Expect()
	testutil.MustMessageContains(testutil.MustCode(resp, 400), "不存在")
}
