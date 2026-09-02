package service

import "mh-sso-svc/internal/audit"

// recordSecurityAudit adapts service metadata to the centralized security
// audit recorder. Authentication events have no MySQL row change to consume.
func (s *AuthService) recordSecurityAudit(operatorID *uint64, account, eventType string, success bool, failReason string, meta RequestMeta) {
	s.securityAudit.Record(audit.SecurityEvent{
		OperatorID: operatorID,
		Account:    account,
		Action:     eventType,
		Success:    success,
		FailReason: failReason,
		IP:         meta.IP,
		RequestID:  meta.RequestID,
	})
}
