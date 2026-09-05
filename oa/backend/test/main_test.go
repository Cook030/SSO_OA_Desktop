package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"permission-system/internal/db_model/query"
	"permission-system/internal/router"
	"permission-system/internal/service/bootstrap"
	"permission-system/internal/utils"
	"permission-system/test/testutil"
)

// TestMain 初始化测试环境并执行全部测试
func TestMain(m *testing.M) {
	testutil.Cfg = testutil.LoadTestConfig()
	testutil.Cfg.Server.Mode = "release"

	utils.InitLogger(testutil.Cfg.Server.Mode, &testutil.Cfg.Log)

	db, closeDB, err := utils.InitDatabase(&testutil.Cfg.MySQL, testutil.Cfg.Server.Mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化数据库失败: %v\n", err)
		os.Exit(1)
	}

	q := query.Use(db)
	if err := bootstrap.InitRBAC(q, &testutil.Cfg.Admin); err != nil {
		fmt.Fprintf(os.Stderr, "初始化 RBAC 失败: %v\n", err)
		os.Exit(1)
	}

	// 启动 mock SSO 服务，并将配置指向它
	mockSSO := newMockSSOServer(q)
	defer mockSSO.Close()
	testutil.Cfg.SSO.BaseURL = mockSSO.URL

	r := router.SetupRouter(db, testutil.Cfg)
	testServer := httptest.NewServer(r)
	defer func() {
		testServer.Close()
		if closeDB != nil {
			closeDB()
		}
	}()

	testutil.BaseURL = testServer.URL + "/api"

	testutil.NoAuthE = testutil.NewClient("")

	testutil.AdminToken = testutil.SSOTokenFor(testutil.Cfg.Admin.Account)
	testutil.E = testutil.NewClient(testutil.AdminToken)

	testutil.GlobalEmployeeID = setupGlobalEmployee()
	testutil.EmployeeToken = testutil.SSOTokenFor(globalEmployeeAccount())
	testutil.EmployeeE = testutil.NewClient(testutil.EmployeeToken)

	code := m.Run()

	cleanupGlobalEmployee()
	testutil.RunCleanup()

	os.Exit(code)
}

// newMockSSOServer 启动模拟 SSO 服务
// token 格式为 "mock-sso-token-{account}" 时返回 200 + 该账号本地用户的数字 id（与 sys_user.id 对齐），否则返回 401
func newMockSSOServer(q *query.Query) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/introspect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		token := ""
		if ck, err := r.Cookie(testutil.CookieAccessTokenName); err == nil {
			token = ck.Value
		}
		account, ok := strings.CutPrefix(token, testutil.MockSSOTokenPrefix)
		if !ok || account == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"code": 401, "msg": "Token 已过期"})
			return
		}

		// 根据账号查出本地用户 id 后返回（与真实 SSO 行为一致）
		user, err := q.SysUser.Where(q.SysUser.Account.Eq(account)).First()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"code": 401, "msg": "用户不存在"})
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"code": 200,
			"msg":  "成功",
			"data": map[string]any{"userId": user.ID, "valid": true},
		})
	})

	// 模拟 SSO 撤销用户会话接口
	mux.HandleFunc("/api/v1/auth/revoke-user-sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]any{"code": 405, "msg": "方法不允许"})
			return
		}

		var reqBody struct {
			UserID int64 `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.UserID <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{"code": 400, "msg": "userId is required"})
			return
		}

		// 本地校验 userId 存在即可
		_, err := q.SysUser.Where(q.SysUser.ID.Eq(reqBody.UserID)).First()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{"code": 400, "msg": "用户不存在"})
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"code": 200,
			"msg":  "success",
			"data": nil,
		})
	})
	return httptest.NewServer(mux)
}



var globalEmployeeEmailPrefix string
var globalAccount string

// setupGlobalEmployee 创建一个全局员工，用于权限不足测试
func setupGlobalEmployee() int64 {
	name, phone, emailPrefix, account := testutil.RandomEmployee()
	globalEmployeeEmailPrefix = emailPrefix
	globalAccount = account

	resp := testutil.E.POST("/employees").WithJSON(map[string]any{
		"name":        name,
		"phone":       phone,
		"emailPrefix": emailPrefix,
		"account":     account,
		"department":  "测试部",
		"platformIds": []int64{},
		"password":    "Test@123456",
	}).Expect()

	data := resp.Status(200).JSON().Object().Value("data").Object()
	return int64(data.Value("id").Number().Raw())
}

// cleanupGlobalEmployee 删除全局员工
func cleanupGlobalEmployee() {
	if testutil.GlobalEmployeeID > 0 {
		testutil.E.DELETE("/employees/{id}").WithPath("id", testutil.GlobalEmployeeID).Expect().Status(200)
	}
}

// globalEmployeeAccount 返回全局员工账号
func globalEmployeeAccount() string {
	return globalAccount
}
