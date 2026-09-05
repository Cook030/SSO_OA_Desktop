// Package rbac 鉴权引擎：基于项目 RBAC 表的权限判定与缓存，
// 与业务服务解耦，可独立替换实现（如后续换 casbin）。
package rbac

import (
	"sort"
	"sync"
	"time"

	"permission-system/internal/repository"
	"permission-system/internal/service/shared"
)

// Subject 鉴权主体：用户 ID + 其当前生效的角色编码列表
type Subject struct {
	UserID int64
	Roles  []string
}

// Enforcer 权限判定器。
//
// 采用 Casbin 风格的 subject / object / action 三元组模型，但策略来源为本项目的 RBAC 表
// （sys_role / sys_permission / sys_role_permission / sys_user_role）而非 casbin_rule 表，
// 判定过程为纯内存匹配。后续若需要更复杂的 matcher（如行级数据权限），
// 可替换为 casbin 实现而不影响调用方。
type Enforcer interface {
	// Enforce 判定主体是否拥有对 object 执行 action 的权限
	Enforce(subject Subject, object, action string) bool
	// LoadRoles 获取用户当前生效的角色编码（带缓存）
	LoadRoles(userID int64) ([]string, error)
	// HasRole 判断用户是否持有指定角色
	HasRole(userID int64, roleCode string) (bool, error)
	// PermissionCodes 获取用户拥有的全部权限编码（含通配符）
	PermissionCodes(userID int64) ([]string, error)
	// ReloadPolicy 重新从数据库加载策略（角色 -> object -> action）
	ReloadPolicy() error
	// InvalidateUser 失效指定用户的角色缓存
	InvalidateUser(userID int64)
	// InvalidatePolicy 策略变更后的失效：重载策略并清空全部用户角色缓存
	InvalidatePolicy() error
}

// 默认缓存时长
const (
	defaultRoleCacheTTL   = 5 * time.Minute
	defaultPolicyCacheTTL = 5 * time.Minute
)

// CachedEnforcer 基于内存策略的 Enforcer 实现（并发安全）
type CachedEnforcer struct {
	permRepo     *repository.PermissionRepository
	userRoleRepo *repository.UserRoleRepository

	roleTTL   time.Duration
	policyTTL time.Duration

	// policy 策略表：roleCode -> object -> actions
	policyMu       sync.RWMutex
	policy         map[string]map[string]map[string]struct{}
	rolePermCodes  map[string][]string
	policyLoadedAt time.Time

	// roleCache 用户角色缓存：userID -> (角色编码, 过期时间)
	roleMu    sync.Mutex
	roleCache map[int64]roleCacheEntry
}

type roleCacheEntry struct {
	roles    []string
	expireAt time.Time
}

// NewCachedEnforcer 创建带缓存的 Enforcer
func NewCachedEnforcer(permRepo *repository.PermissionRepository, userRoleRepo *repository.UserRoleRepository) *CachedEnforcer {
	return &CachedEnforcer{
		permRepo:      permRepo,
		userRoleRepo:  userRoleRepo,
		roleTTL:       defaultRoleCacheTTL,
		policyTTL:     defaultPolicyCacheTTL,
		policy:        make(map[string]map[string]map[string]struct{}),
		rolePermCodes: make(map[string][]string),
		roleCache:     make(map[int64]roleCacheEntry),
	}
}

// Enforce 判定主体是否拥有对 object 执行 action 的权限。
// 匹配优先级：精确 object:action > object 通配(<object>:*) > object 全通配(*:*)
func (e *CachedEnforcer) Enforce(subject Subject, object, action string) bool {
	if len(subject.Roles) == 0 {
		return false
	}
	if err := e.ensurePolicy(); err != nil {
		return false
	}

	e.policyMu.RLock()
	defer e.policyMu.RUnlock()

	for _, role := range subject.Roles {
		if matchPolicy(e.policy[role], object, action) {
			return true
		}
	}
	return false
}

// matchPolicy 在单个角色的权限集合内匹配 object/action
func matchPolicy(perms map[string]map[string]struct{}, object, action string) bool {
	if len(perms) == 0 {
		return false
	}

	// 精确资源
	if actions, ok := perms[object]; ok {
		if _, ok := actions[action]; ok {
			return true
		}
		if _, ok := actions[shared.Wildcard]; ok { // <object>:*
			return true
		}
	}

	// 资源通配：* 可匹配任意 object
	if actions, ok := perms[shared.Wildcard]; ok {
		if _, ok := actions[action]; ok { // *:<action>
			return true
		}
		if _, ok := actions[shared.Wildcard]; ok { // *:*
			return true
		}
	}
	return false
}

// LoadRoles 获取用户当前生效的角色编码（带缓存）
func (e *CachedEnforcer) LoadRoles(userID int64) ([]string, error) {
	if roles, ok := e.cachedRoles(userID); ok {
		return roles, nil
	}

	roles, err := e.userRoleRepo.FindRoleCodesByUserID(userID)
	if err != nil {
		return nil, err
	}

	e.roleMu.Lock()
	e.roleCache[userID] = roleCacheEntry{roles: roles, expireAt: time.Now().Add(e.roleTTL)}
	e.roleMu.Unlock()

	return roles, nil
}

// HasRole 判断用户是否持有指定角色
func (e *CachedEnforcer) HasRole(userID int64, roleCode string) (bool, error) {
	roles, err := e.LoadRoles(userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == roleCode {
			return true, nil
		}
	}
	return false, nil
}

// PermissionCodes 获取用户拥有的全部权限编码（去重且有序）
func (e *CachedEnforcer) PermissionCodes(userID int64) ([]string, error) {
	if err := e.ensurePolicy(); err != nil {
		return nil, err
	}
	roles, err := e.LoadRoles(userID)
	if err != nil {
		return nil, err
	}

	e.policyMu.RLock()
	defer e.policyMu.RUnlock()

	set := make(map[string]struct{})
	for _, role := range roles {
		for _, code := range e.rolePermCodes[role] {
			set[code] = struct{}{}
		}
	}

	codes := make([]string, 0, len(set))
	for code := range set {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes, nil
}

// ReloadPolicy 重新从数据库加载策略
func (e *CachedEnforcer) ReloadPolicy() error {
	rows, err := e.permRepo.ListPolicy()
	if err != nil {
		return err
	}

	policy := make(map[string]map[string]map[string]struct{}, len(rows))
	rolePermCodes := make(map[string][]string, len(rows))
	for _, row := range rows {
		object, action := shared.SplitPermissionCode(row.PermCode)

		objects, ok := policy[row.RoleCode]
		if !ok {
			objects = make(map[string]map[string]struct{})
			policy[row.RoleCode] = objects
		}
		actions, ok := objects[object]
		if !ok {
			actions = make(map[string]struct{})
			objects[object] = actions
		}
		actions[action] = struct{}{}

		rolePermCodes[row.RoleCode] = append(rolePermCodes[row.RoleCode], row.PermCode)
	}

	e.policyMu.Lock()
	e.policy = policy
	e.rolePermCodes = rolePermCodes
	e.policyLoadedAt = time.Now()
	e.policyMu.Unlock()
	return nil
}

// InvalidateUser 失效指定用户的角色缓存
func (e *CachedEnforcer) InvalidateUser(userID int64) {
	e.roleMu.Lock()
	delete(e.roleCache, userID)
	e.roleMu.Unlock()
}

// InvalidatePolicy 策略变更后调用：重载策略，并清空全部用户角色缓存
// （角色被禁用/删除时，用户角色归属同时发生变化，故一并清空）
func (e *CachedEnforcer) InvalidatePolicy() error {
	e.roleMu.Lock()
	e.roleCache = make(map[int64]roleCacheEntry)
	e.roleMu.Unlock()

	e.policyMu.Lock()
	e.policyLoadedAt = time.Time{}
	e.policyMu.Unlock()

	return e.ReloadPolicy()
}

// ensurePolicy 策略过期时自动重载，避免直接改库后长期不生效
func (e *CachedEnforcer) ensurePolicy() error {
	e.policyMu.RLock()
	fresh := !e.policyLoadedAt.IsZero() && time.Since(e.policyLoadedAt) < e.policyTTL
	e.policyMu.RUnlock()

	if fresh {
		return nil
	}
	return e.ReloadPolicy()
}

// cachedRoles 读取未过期的角色缓存
func (e *CachedEnforcer) cachedRoles(userID int64) ([]string, bool) {
	e.roleMu.Lock()
	defer e.roleMu.Unlock()

	entry, ok := e.roleCache[userID]
	if !ok || time.Now().After(entry.expireAt) {
		if ok {
			delete(e.roleCache, userID)
		}
		return nil, false
	}
	return entry.roles, true
}
