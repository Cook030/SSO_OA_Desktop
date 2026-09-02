package audit

import (
	"strings"

	"mh-sso-svc/internal/repository"

	"go.uber.org/zap"
)

// SecurityEvent describes an authentication-security event. These events are
// intentionally written directly because Redis-backed login and token changes
// do not generate MySQL binlog records.
type SecurityEvent struct {
	OperatorID *uint64
	Account    string
	Action     string
	Success    bool
	FailReason string
	IP         string
	RequestID  string
}

// SecurityRecorder is the sole persistence boundary for SSO security audits.
type SecurityRecorder struct {
	repo *repository.AuditRepository
	log  *zap.Logger
}

func NewSecurityRecorder(repo *repository.AuditRepository, log *zap.Logger) *SecurityRecorder {
	return &SecurityRecorder{repo: repo, log: log}
}

// Record persists an audit event best-effort, never interrupting authentication
// flows if the audit table is temporarily unavailable.
func (r *SecurityRecorder) Record(event SecurityEvent) {
	record := &repository.AuditRecord{
		OperatorID: event.OperatorID,
		Account:    nilIfEmpty(truncate(event.Account, 64)),
		Action:     event.Action,
		Success:    event.Success,
		FailReason: nilIfEmpty(event.FailReason),
		IP:         nilIfEmpty(event.IP),
		RequestID:  nilIfEmpty(event.RequestID),
	}
	if err := r.repo.Create(record); err != nil {
		r.log.Error("写入安全审计日志失败", zap.String("event_type", event.Action), zap.Error(err))
	}
}

func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func truncate(value string, maxLength int) string {
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}
	return strings.TrimSpace(value[:maxLength])
}
