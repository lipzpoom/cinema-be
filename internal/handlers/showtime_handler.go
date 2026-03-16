package handlers

import (
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type ShowtimeHandler struct {
	Service *services.ShowtimeService
}

func NewShowtimeHandler(service *services.ShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{Service: service}
}

// Create godoc
// @Summary Create a new showtime
// @Description Add a new showtime to the system (Admin only)
// @Tags Showtimes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param showtime body models.Showtime true "Showtime Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/showtimes [post]
func (h *ShowtimeHandler) Create(c *gin.Context) {
	var item models.Showtime
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := h.Service.Create(c.Request.Context(), &item); err != nil {
		if errors.Is(err, services.ErrShowtimeOverlap) {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflicting showtime in this theater"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Created successfully"})
}

// GetAll godoc
// @Summary Get all showtimes
// @Description Fetch a list of all showtimes
// @Tags Showtimes
// @Produce json
// @Success 200 {array} models.Showtime
// @Failure 500 {object} map[string]interface{}
// @Router /api/showtimes [get]
func (h *ShowtimeHandler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary Get a showtime by ID
// @Description Fetch a single showtime by its ID
// @Tags Showtimes
// @Produce json
// @Param id path string true "Showtime ID"
// @Success 200 {object} models.Showtime
// @Failure 404 {object} map[string]interface{}
// @Router /api/showtimes/{id} [get]
func (h *ShowtimeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary Update a showtime
// @Description Update showtime details by ID (Admin only)
// @Tags Showtimes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Showtime ID"
// @Param updateData body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/showtimes/{id} [put]
func (h *ShowtimeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData["updated_at"] = time.Now()

	if err := h.Service.Update(c.Request.Context(), id, updateData); err != nil {
		if errors.Is(err, services.ErrShowtimeOverlap) {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflicting showtime in this theater"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}

// Delete godoc
// @Summary Delete a showtime
// @Description Delete a showtime from the system by ID (Admin only)
// @Tags Showtimes
// @Security BearerAuth
// @Produce json
// @Param id path string true "Showtime ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/showtimes/{id} [delete]
func (h *ShowtimeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// GetByMovieID godoc
// @Summary Get showtimes by Movie ID
// @Description Fetch a list of showtimes for a specific movie
// @Tags Movies
// @Produce json
// @Param id path string true "Movie ID"
// @Success 200 {array} models.Showtime
// @Failure 500 {object} map[string]interface{}
// @Router /api/movies/{id}/showtimes [get]
func (h *ShowtimeHandler) GetByMovieID(c *gin.Context) {
	movieID := c.Param("id")
	items, err := h.Service.GetByMovieID(c.Request.Context(), movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetNowShowing godoc
// @Summary Get now showing showtimes
// @Description Fetch showtimes for a specific date (YYYY-MM-DD or all if absent)
// @Tags Showtimes
// @Produce json
// @Param date query string false "Date in YYYY-MM-DD"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/showtimes/now-showing [get]
func (h *ShowtimeHandler) GetNowShowing(c *gin.Context) {
	date := c.Query("date") // Format: YYYY-MM-DD
	items, err := h.Service.GetNowShowing(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return empty array instead of null if no items
	if items == nil {
		items = []bson.M{}
	}
	c.JSON(http.StatusOK, items)
}
