package handlers

import (
	"fmt"
	"gin-quickstart/internal/events"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type SeatHandler struct {
	Service *services.SeatService
}

func NewSeatHandler(service *services.SeatService) *SeatHandler {
	return &SeatHandler{Service: service}
}

// Create godoc
// @Summary Create a new seat
// @Description Add a new seat to the system (Admin only)
// @Tags Seats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param seat body models.Seat true "Seat Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/seats [post]
func (h *SeatHandler) Create(c *gin.Context) {
	var item models.Seat
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.Create(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add Audit Log
	userID, _ := c.Get("user_id")
	_ = events.PublishAuditLog(events.AuditLogPayload{
		Event:     "CREATE_SEAT",
		UserID:    fmt.Sprintf("%v", userID),
		Value:     item.SeatNumber, 
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Created successfully"})
}

// GetAll godoc
// @Summary Get all seats
// @Description Fetch a list of all seats
// @Tags Seats
// @Produce json
// @Success 200 {array} models.Seat
// @Failure 500 {object} map[string]interface{}
// @Router /api/seats [get]
func (h *SeatHandler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary Get a seat by ID
// @Description Fetch a single seat by its ID
// @Tags Seats
// @Produce json
// @Param id path string true "Seat ID"
// @Success 200 {object} models.Seat
// @Failure 404 {object} map[string]interface{}
// @Router /api/seats/{id} [get]
func (h *SeatHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary Update a seat
// @Description Update seat details by ID (Admin only)
// @Tags Seats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Seat ID"
// @Param updateData body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/seats/{id} [put]
func (h *SeatHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.Update(c.Request.Context(), id, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}

// Delete godoc
// @Summary Delete a seat
// @Description Delete a seat from the system by ID (Admin only)
// @Tags Seats
// @Security BearerAuth
// @Produce json
// @Param id path string true "Seat ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/seats/{id} [delete]
func (h *SeatHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == "cannot delete seat: it is associated with an existing booking" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Add Audit Log
	userID, _ := c.Get("user_id")
	_ = events.PublishAuditLog(events.AuditLogPayload{
		Event:     "DELETE_SEAT",
		UserID:    fmt.Sprintf("%v", userID),
		Value:     id, // using id as seat target for reference
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}
