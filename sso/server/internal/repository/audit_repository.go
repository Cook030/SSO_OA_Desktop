package repository

import (
	"mh-sso-svc/internal/db_model"
	"mh-sso-svc/internal/db_model/query"
)

// AuditRepository 业务审计日志表（sys_audit_log）数据访问
type AuditRepository struct {
	q *query.Query
}

// NewAuditRepository 创建审计 Repository
func NewAuditRepository(q *query.Query) *AuditRepository {
	return &AuditRepository{q: q}
}

// AuditRecord SSO 侧审计记录入参（兼容原 SsoLoginAudit 语义）
type AuditRecord struct {
	UserID     *uint64
	Account    *string
	Action     string
	Success    bool
	FailReason *string
	IP         *string
	UserAgent  *string
	RequestID  *string
}

// Create 新增审计记录（失败由调用方决定是否忽略，不阻断主流程）
func (r *AuditRepository) Create(record *AuditRecord) error {
	result := int32(1)
	if !record.Success {
		result = 0
	}
	return r.q.SysAuditLog.Create(&db_model.SysAuditLog{
		UserID:       record.UserID,
		Action:       record.Action,
		TargetType:   strPtr("user"),
		TargetID:     accountID(record.Account),
		IP:           record.IP,
		RequestID:    record.RequestID,
		Result:       result,
		ErrorMessage: record.FailReason,
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func accountID(account *string) *string {
	if account == nil || *account == "" {
		return nil
	}
	return account
}
