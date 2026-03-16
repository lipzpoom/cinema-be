package handlers

import (
	"net/http"

	"gin-quickstart/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	Service *services.AuditLogService
}

func NewAuditLogHandler(service *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		Service: service,
	}
}

// GetAllAuditLogs godoc
// @Summary      Get all audit logs
// @Description  Retrieves all audit logs in descending order.
// @Tags         Audit Logs
// @Produce      json
// @Success      200  {array}   models.AuditLog
// @Router       /api/auditlogs [get]
func (h *AuditLogHandler) GetAllAuditLogs(c *gin.Context) {
	logs, err := h.Service.GetAllAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
