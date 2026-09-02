package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

const defaultAdminPassword = "admin123"

type App struct {
	ctx       context.Context
	db        *sql.DB
	dataDir   string
	config    WorkbenchConfig
	session   *Session
	sessionMu sync.RWMutex
}

type WorkbenchConfig struct {
	App struct {
		Name string `yaml:"name"`
		Mode string `yaml:"mode"`
	} `yaml:"app"`
	Auth struct {
		Provider string `yaml:"provider"`
	} `yaml:"auth"`
	Executors struct {
		DefaultAlias string `yaml:"default_alias"`
	} `yaml:"executors"`
	OA struct {
		BaseURL        string `yaml:"base_url"`
		AccessTokenEnv string `yaml:"access_token_env"`
		CookieName     string `yaml:"cookie_name"`
		Mode           string `yaml:"mode"`
	} `yaml:"oa"`
}

type Session struct {
	UserID      int64  `json:"userId"`
	Account     string `json:"account"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}
type Skill struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Owner        string `json:"owner"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Enabled      bool   `json:"enabled"`
	Tags         string `json:"tags"`
	UpdatedAt    string `json:"updatedAt"`
	SnapshotPath string `json:"-"`
}
type Run struct {
	ID            int64  `json:"id"`
	SkillName     string `json:"skillName"`
	SkillVersion  string `json:"skillVersion"`
	Operator      string `json:"operator"`
	DataLevel     string `json:"dataLevel"`
	ExecutorAlias string `json:"executorAlias"`
	Status        string `json:"status"`
	Output        string `json:"output"`
	ErrorMessage  string `json:"errorMessage"`
	CreatedAt     string `json:"createdAt"`
}
type CreateSkillRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}
type StartRunRequest struct {
	SkillID      int64  `json:"skillId"`
	DataLevel    string `json:"dataLevel"`
	InputSummary string `json:"inputSummary"`
}
type ImportResult struct {
	Imported bool   `json:"imported"`
	Message  string `json:"message"`
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.initialise(); err != nil {
		runtime.LogErrorf(ctx, "初始化工作台失败: %v", err)
	}
}

func (a *App) initialise() error {
	base, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	a.dataDir = filepath.Join(base, "MapleHaze", "AIWorkbench")
	for _, dir := range []string{"data", "skills", "artifacts", "imports", "config"} {
		if err := os.MkdirAll(filepath.Join(a.dataDir, dir), 0o700); err != nil {
			return err
		}
	}
	if err := a.loadConfig(); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", filepath.Join(a.dataDir, "data", "workbench.db"))
	if err != nil {
		return err
	}
	a.db = db
	if err := a.migrate(); err != nil {
		return err
	}
	if err := a.seedAdmin(); err != nil {
		return err
	}
	return a.seedExampleSkills()
}

func (a *App) loadConfig() error {
	configPath := filepath.Join(a.dataDir, "config", "workbench.yaml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		content := "app:\n  name: 智工工作台\n  mode: standalone\nauth:\n  provider: local\nexecutors:\n  default_alias: local-agent\noa:\n  base_url: http://127.0.0.1:8080\n  access_token_env: OA_ACCESS_TOKEN\n  cookie_name: mh_sso2_access_token\n  mode: dry-run\n"
		if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &a.config); err != nil {
		return err
	}
	if a.config.App.Name == "" {
		a.config.App.Name = "智工工作台"
	}
	if a.config.Executors.DefaultAlias == "" {
		a.config.Executors.DefaultAlias = "local-agent"
	}
	if a.config.OA.CookieName == "" {
		a.config.OA.CookieName = "mh_sso2_access_token"
	}
	if a.config.OA.Mode == "" {
		a.config.OA.Mode = "dry-run"
	}
	return nil
}

func (a *App) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, account TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, display_name TEXT NOT NULL, role TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS skills (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT NOT NULL, owner TEXT NOT NULL, tags TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS skill_versions (id INTEGER PRIMARY KEY AUTOINCREMENT, skill_id INTEGER NOT NULL, version TEXT NOT NULL, status TEXT NOT NULL, snapshot_path TEXT NOT NULL, content_hash TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(skill_id, version))`,
		`CREATE TABLE IF NOT EXISTS runs (id INTEGER PRIMARY KEY AUTOINCREMENT, skill_id INTEGER NOT NULL, version TEXT NOT NULL, operator TEXT NOT NULL, data_level TEXT NOT NULL, executor_alias TEXT NOT NULL, status TEXT NOT NULL, input_summary TEXT NOT NULL, output TEXT NOT NULL DEFAULT '', error_message TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY AUTOINCREMENT, actor TEXT NOT NULL, event_type TEXT NOT NULL, resource_type TEXT NOT NULL, resource_id TEXT NOT NULL, details TEXT NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS conversations (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY AUTOINCREMENT, conversation_id INTEGER NOT NULL, role TEXT NOT NULL, content TEXT NOT NULL, created_at TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := a.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) seedAdmin() error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.db.Exec(`INSERT INTO users(account,password_hash,display_name,role,created_at) VALUES(?,?,?,?,?)`, "admin", string(hash), "本地管理员", "admin", now())
	return err
}
func (a *App) seedExampleSkills() error {
	for _, example := range []CreateSkillRequest{{"data-analyze-tabular", "分析显式上传的 CSV/XLSX 表格并输出结构化报告。", "data-analysis"}, {"debug-analyze-log", "基于日志文本与问题描述整理诊断证据和下一步建议。", "debug"}} {
		var exists int
		if err := a.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE name=?`, example.Name).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			if _, err := a.createSkill("admin", example, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) Login(account, password string) (*Session, error) {
	if a.db == nil {
		return nil, errors.New("工作台正在初始化")
	}
	var user Session
	var hash string
	var enabled int
	err := a.db.QueryRow(`SELECT id,account,display_name,role,password_hash,enabled FROM users WHERE account=?`, strings.TrimSpace(account)).Scan(&user.UserID, &user.Account, &user.DisplayName, &user.Role, &hash, &enabled)
	if err != nil || enabled == 0 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errors.New("账号或密码错误")
	}
	a.sessionMu.Lock()
	a.session = &user
	a.sessionMu.Unlock()
	return &user, nil
}
func (a *App) CurrentSession() *Session {
	a.sessionMu.RLock()
	defer a.sessionMu.RUnlock()
	return a.session
}
func (a *App) Logout()         { a.sessionMu.Lock(); a.session = nil; a.sessionMu.Unlock() }
func (a *App) AppName() string { return a.config.App.Name }

func (a *App) ListSkills() ([]Skill, error) {
	rows, err := a.db.Query(`SELECT s.id,s.name,s.description,s.owner,s.tags,s.enabled,s.updated_at,v.version,v.status,v.snapshot_path FROM skills s JOIN skill_versions v ON v.skill_id=s.id WHERE v.id=(SELECT id FROM skill_versions WHERE skill_id=s.id ORDER BY id DESC LIMIT 1) ORDER BY s.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Skill{}
	for rows.Next() {
		var item Skill
		var enabled int
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Owner, &item.Tags, &enabled, &item.UpdatedAt, &item.Version, &item.Status, &item.SnapshotPath); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		items = append(items, item)
	}
	return items, rows.Err()
}
func (a *App) CreateSkill(req CreateSkillRequest) (*Skill, error) {
	actor, err := a.requireRole("admin", "maintainer")
	if err != nil {
		return nil, err
	}
	return a.createSkill(actor.Account, req, false)
}
func (a *App) createSkill(owner string, req CreateSkillRequest, published bool) (*Skill, error) {
	if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(req.Name) {
		return nil, errors.New("Skill 名称只能使用小写字母、数字和连字符")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, errors.New("请填写 Skill 描述")
	}
	created := now()
	result, err := a.db.Exec(`INSERT INTO skills(name,description,owner,tags,enabled,created_at,updated_at) VALUES(?,?,?,?,1,?,?)`, req.Name, req.Description, owner, req.Tags, created, created)
	if err != nil {
		return nil, friendlyDBError(err)
	}
	id, _ := result.LastInsertId()
	status := "draft"
	if published {
		status = "published"
	}
	snapshot, hash, err := a.writeSnapshot(req.Name, "0.1.0", req.Description, req.Tags)
	if err != nil {
		return nil, err
	}
	if _, err := a.db.Exec(`INSERT INTO skill_versions(skill_id,version,status,snapshot_path,content_hash,created_at) VALUES(?,?,?,?,?,?)`, id, "0.1.0", status, snapshot, hash, created); err != nil {
		return nil, err
	}
	a.recordEvent(owner, "skill.created", "skill", fmt.Sprint(id), req.Name)
	return &Skill{ID: id, Name: req.Name, Description: req.Description, Owner: owner, Version: "0.1.0", Status: status, Enabled: true, Tags: req.Tags, UpdatedAt: created}, nil
}
func (a *App) PublishSkill(skillID int64) error {
	actor, err := a.requireRole("admin")
	if err != nil {
		return err
	}
	var versionID int64
	err = a.db.QueryRow(`SELECT id FROM skill_versions WHERE skill_id=? ORDER BY id DESC LIMIT 1`, skillID).Scan(&versionID)
	if err != nil {
		return errors.New("未找到 Skill")
	}
	if _, err = a.db.Exec(`UPDATE skill_versions SET status='published' WHERE id=?`, versionID); err != nil {
		return err
	}
	if _, err = a.db.Exec(`UPDATE skills SET updated_at=? WHERE id=?`, now(), skillID); err != nil {
		return err
	}
	a.recordEvent(actor.Account, "skill.published", "skill", fmt.Sprint(skillID), "")
	return nil
}
func (a *App) ToggleSkill(skillID int64, enabled bool) error {
	actor, err := a.requireRole("admin")
	if err != nil {
		return err
	}
	if _, err = a.db.Exec(`UPDATE skills SET enabled=?,updated_at=? WHERE id=?`, boolToInt(enabled), now(), skillID); err != nil {
		return err
	}
	a.recordEvent(actor.Account, "skill.toggled", "skill", fmt.Sprint(skillID), fmt.Sprint(enabled))
	return nil
}

func (a *App) StartRun(req StartRunRequest) (*Run, error) {
	actor, err := a.requireRole("admin", "maintainer", "user")
	if err != nil {
		return nil, err
	}
	if !validDataLevel(req.DataLevel) {
		return nil, errors.New("数据等级不合法")
	}
	var skill Skill
	var enabled int
	err = a.db.QueryRow(`SELECT s.name,s.owner,s.enabled,v.version,v.status FROM skills s JOIN skill_versions v ON v.skill_id=s.id WHERE s.id=? ORDER BY v.id DESC LIMIT 1`, req.SkillID).Scan(&skill.Name, &skill.Owner, &enabled, &skill.Version, &skill.Status)
	if err != nil {
		return nil, errors.New("未找到 Skill")
	}
	if enabled == 0 || skill.Status != "published" {
		return nil, errors.New("Skill 未发布或已停用")
	}
	created := now()
	result, err := a.db.Exec(`INSERT INTO runs(skill_id,version,operator,data_level,executor_alias,status,input_summary,created_at,updated_at) VALUES(?,?,?,?,?,'queued',?,?,?)`, req.SkillID, skill.Version, actor.Account, req.DataLevel, a.config.Executors.DefaultAlias, req.InputSummary, created, created)
	if err != nil {
		return nil, err
	}
	runID, _ := result.LastInsertId()
	run := &Run{ID: runID, SkillName: skill.Name, SkillVersion: skill.Version, Operator: actor.Account, DataLevel: req.DataLevel, ExecutorAlias: a.config.Executors.DefaultAlias, Status: "queued", CreatedAt: created}
	go a.executeMock(runID, skill.Name, req.InputSummary)
	return run, nil
}
func (a *App) executeMock(runID int64, skillName, input string) {
	_, _ = a.db.Exec(`UPDATE runs SET status='running',updated_at=? WHERE id=?`, now(), runID)
	time.Sleep(900 * time.Millisecond)
	output := fmt.Sprintf("# %s 执行报告\n\n## 执行摘要\n\n已使用 Mock 执行器完成 MVP 调用。\n\n## 输入摘要\n\n%s\n\n## 下一步\n\n接入 OpenAI 兼容执行器后，将在此生成真实 Skill 产物。", skillName, emptyAs(input, "未提供"))
	artifactDir := filepath.Join(a.dataDir, "artifacts", fmt.Sprint(runID))
	_ = os.MkdirAll(artifactDir, 0o700)
	_ = os.WriteFile(filepath.Join(artifactDir, "report.md"), []byte(output), 0o600)
	_, _ = a.db.Exec(`UPDATE runs SET status='succeeded',output=?,updated_at=? WHERE id=?`, output, now(), runID)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "run:updated", runID)
	}
}
func (a *App) ListRuns() ([]Run, error) {
	rows, err := a.db.Query(`SELECT r.id,s.name,r.version,r.operator,r.data_level,r.executor_alias,r.status,r.output,r.error_message,r.created_at FROM runs r JOIN skills s ON s.id=r.skill_id ORDER BY r.id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Run{}
	for rows.Next() {
		var item Run
		if err := rows.Scan(&item.ID, &item.SkillName, &item.SkillVersion, &item.Operator, &item.DataLevel, &item.ExecutorAlias, &item.Status, &item.Output, &item.ErrorMessage, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *App) ImportSkillPackage() (*ImportResult, error) {
	actor, err := a.requireRole("admin", "maintainer")
	if err != nil {
		return nil, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "导入 Skill 包", Filters: []runtime.FileFilter{{DisplayName: "ZIP 包", Pattern: "*.zip"}}})
	if err != nil || path == "" {
		return &ImportResult{Imported: false, Message: "已取消导入"}, err
	}
	meta, description, tags, err := inspectSkillZip(path)
	if err != nil {
		return nil, err
	}
	item, err := a.createSkill(actor.Account, CreateSkillRequest{Name: meta.Name, Description: description, Tags: tags}, false)
	if err != nil {
		return nil, err
	}
	return &ImportResult{Imported: true, Message: fmt.Sprintf("已导入 %s %s（草稿）", item.Name, item.Version)}, nil
}

type skillMeta struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`
}

func inspectSkillZip(path string) (skillMeta, string, string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return skillMeta{}, "", "", errors.New("无法读取 ZIP 包")
	}
	defer r.Close()
	if len(r.File) > 200 {
		return skillMeta{}, "", "", errors.New("Skill 包文件数量不能超过 200")
	}
	var raw string
	for _, f := range r.File {
		if strings.Contains(f.Name, "..") || strings.HasPrefix(f.Name, "/") {
			return skillMeta{}, "", "", errors.New("Skill 包包含非法路径")
		}
		if strings.EqualFold(filepath.Base(f.Name), "SKILL.md") {
			rc, e := f.Open()
			if e != nil {
				return skillMeta{}, "", "", e
			}
			b, _ := io.ReadAll(io.LimitReader(rc, 1024*1024))
			rc.Close()
			raw = string(b)
		}
	}
	if raw == "" {
		return skillMeta{}, "", "", errors.New("Skill 包缺少 SKILL.md")
	}
	parts := strings.SplitN(raw, "---", 3)
	if len(parts) < 3 {
		return skillMeta{}, "", "", errors.New("SKILL.md 缺少 YAML Front Matter")
	}
	var meta skillMeta
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return skillMeta{}, "", "", errors.New("SKILL.md 元数据无法解析")
	}
	if meta.Name == "" {
		return skillMeta{}, "", "", errors.New("SKILL.md 缺少 name")
	}
	desc := strings.TrimSpace(meta.Description)
	if desc == "" {
		desc = strings.TrimSpace(parts[2])
	}
	return meta, desc, strings.Join(meta.Tags, ","), nil
}
func (a *App) writeSnapshot(name, version, description, tags string) (string, string, error) {
	dir := filepath.Join(a.dataDir, "skills", name, "versions", version)
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o700); err != nil {
		return "", "", err
	}
	content := fmt.Sprintf("---\nname: %s\ndescription: %s\nversion: %s\nowner: admin\ntags: [%s]\ninput_schema:\n  type: object\noutput_contract:\n  type: markdown\nmodel_alias: mock-default\nrequired_tools: []\nmax_data_level: sensitive\n---\n\n# 执行说明\n\n这是由工作台创建的 MVP Skill 草稿。\n", name, description, version, tags)
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", "", err
	}
	_ = os.WriteFile(filepath.Join(dir, "tests", "cases.yaml"), []byte("cases:\n  - name: happy-path\n    required: true\n    expect:\n      status: succeeded\n"), 0o600)
	sum := sha256.Sum256([]byte(content))
	return dir, hex.EncodeToString(sum[:]), nil
}
func (a *App) requireRole(roles ...string) (*Session, error) {
	s := a.CurrentSession()
	if s == nil {
		return nil, errors.New("请先登录")
	}
	for _, role := range roles {
		if s.Role == role {
			return s, nil
		}
	}
	return nil, errors.New("当前账号没有该操作权限")
}
func (a *App) recordEvent(actor, eventType, resourceType, resourceID, details string) {
	_, _ = a.db.Exec(`INSERT INTO events(actor,event_type,resource_type,resource_id,details,created_at) VALUES(?,?,?,?,?,?)`, actor, eventType, resourceType, resourceID, details, now())
}
func now() string { return time.Now().Format(time.RFC3339) }
func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func validDataLevel(v string) bool {
	for _, level := range []string{"public", "internal", "sensitive", "strictly_restricted"} {
		if v == level {
			return true
		}
	}
	return false
}
func emptyAs(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
func friendlyDBError(err error) error {
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return errors.New("Skill 名称已存在")
	}
	return err
}
