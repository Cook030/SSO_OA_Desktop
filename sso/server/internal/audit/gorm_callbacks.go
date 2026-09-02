package audit

import "gorm.io/gorm"

var auditedTables = map[string]struct{}{
	"sys_user":            {},
	"sys_platform":        {},
	"sys_role":            {},
	"sys_permission":      {},
	"sys_user_role":       {},
	"sys_role_permission": {},
	"sys_user_platform":   {},
}

// RegisterGORMCallbacks installs the provenance callback once when the DB is
// initialized. It only handles tables whose audit columns have been migrated.
func RegisterGORMCallbacks(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gorm:create").Register("audit:populate-create-provenance", populateCreateProvenance); err != nil {
		return err
	}
	return db.Callback().Update().Before("gorm:update").Register("audit:populate-update-provenance", populateUpdateProvenance)
}

func populateCreateProvenance(tx *gorm.DB) {
	populateProvenance(tx, true)
}

// populateUpdateProvenance enriches writes using data carried by the request
// context. GORM Gen's update methods use maps, so no service or repository
// needs to assign audit fields explicitly.
func populateUpdateProvenance(tx *gorm.DB) {
	populateProvenance(tx, false)
}

func populateProvenance(tx *gorm.DB, creating bool) {
	if !isAuditedTable(tx) {
		return
	}

	provenance := FromContext(tx.Statement.Context)
	if provenance.OperatorID == 0 && provenance.RequestID == "" {
		return
	}

	if values, ok := tx.Statement.Dest.(map[string]interface{}); ok {
		if creating && provenance.OperatorID != 0 {
			values["created_by"] = provenance.OperatorID
		}
		if provenance.OperatorID != 0 {
			values["updated_by"] = provenance.OperatorID
		}
		if provenance.RequestID != "" {
			values["request_id"] = provenance.RequestID
		}
		return
	}

	if tx.Statement.Schema == nil {
		return
	}
	if creating && provenance.OperatorID != 0 && tx.Statement.Schema.LookUpField("created_by") != nil {
		tx.Statement.SetColumn("created_by", provenance.OperatorID)
	}
	if provenance.OperatorID != 0 && tx.Statement.Schema.LookUpField("updated_by") != nil {
		tx.Statement.SetColumn("updated_by", provenance.OperatorID)
	}
	if provenance.RequestID != "" && tx.Statement.Schema.LookUpField("request_id") != nil {
		tx.Statement.SetColumn("request_id", provenance.RequestID)
	}
}

func isAuditedTable(tx *gorm.DB) bool {
	if tx == nil || tx.Statement == nil {
		return false
	}
	table := tx.Statement.Table
	if table == "" && tx.Statement.Schema != nil {
		table = tx.Statement.Schema.Table
	}
	_, ok := auditedTables[table]
	return ok
}
