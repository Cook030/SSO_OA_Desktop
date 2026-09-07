package mapper

import "strings"

func (m *Mapper) actionName(table, typ string) string {
	tm := m.tables[table]
	for event, action := range tm.Actions {
		if strings.EqualFold(event, typ) && action != "" {
			return action
		}
	}
	prefix := tm.TargetType
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
