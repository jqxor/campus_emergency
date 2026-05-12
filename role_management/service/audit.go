package service

import "context"

type AuditLogService interface {
	LogAction(ctx context.Context, action, description string, metadata map[string]interface{}) error
}

type NoopAuditLogService struct{}

func (NoopAuditLogService) LogAction(ctx context.Context, action, description string, metadata map[string]interface{}) error {
	return nil
}
