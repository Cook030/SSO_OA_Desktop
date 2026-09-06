package mapper

// 关联表使用更贴合业务的动作语义。
var joinActions = map[string]map[string]string{
	"sys_user_role":       {"INSERT": "user:assign-role", "UPDATE": "user:update", "DELETE": "user:unassign-role"},
	"sys_role_permission": {"INSERT": "role:assign-permission", "UPDATE": "role:update", "DELETE": "role:unassign-permission"},
	"sys_user_platform":   {"INSERT": "user:assign-platform", "UPDATE": "user:update", "DELETE": "user:unassign-platform"},
}

func (m *Mapper) actionName(table, typ string) string {
	if action, ok := joinActions[table][typ]; ok {
		return action
	}
	prefix := m.tables[table].TargetType
	switch typ {
	case "INSERT":
		return prefix + ":create"
	case "UPDATE":
		return prefix + ":update"
	case "DELETE":
		return prefix + ":delete"
	default:
		return prefix + ":unknown"
	}
}
