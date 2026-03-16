package services

import (
	"context"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"
)

type AuditLogService struct {
	Repo *repositories.AuditLogRepository
}

func NewAuditLogService(repo *repositories.AuditLogRepository) *AuditLogService {
	return &AuditLogService{Repo: repo}
}

func (s *AuditLogService) GetAllAuditLogs(ctx context.Context) ([]models.AuditLog, error) {
	return s.Repo.GetAll(ctx)
}
