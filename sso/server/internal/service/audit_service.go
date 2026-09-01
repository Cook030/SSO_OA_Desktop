package service

import (
	"mh-sso-svc/internal/repository"
	"mh-sso-svc/internal/utils"

	"go.uber.org/zap"
)

// recordAudit 写入审计日志（失败只记错误日志，不阻断主流程）
func (s *AuthService) recordAudit(userID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta) {
	record := &repository.AuditRecord{
		UserID:     userID,
		Account:    utils.EmptyToNil(utils.Truncate(account, 64)),
		Action:     eventType,
		Success:    success,
		FailReason: utils.EmptyToNil(failReason),
		IP:         utils.EmptyToNil(meta.IP),
		UserAgent:  utils.EmptyToNil(meta.UserAgent),
		RequestID:  utils.EmptyToNil(meta.RequestID),
	}
	if err := s.auditRepo.Create(record); err != nil {
		s.log.Error("写入审计日志失败", zap.String("event_type", eventType), zap.Error(err))
	}
}
