package handlers

import (
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/services"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovieHandler struct {
	Service *services.MovieService
}

func NewMovieHandler(service *services.MovieService) *MovieHandler {
	return &MovieHandler{Service: service}
}

// Create godoc
// @Summary Create a new movie
// @Description Add a new movie to the system (Admin only)
// @Tags Movies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param movie body models.Movie true "Movie Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/movies [post]
func (h *MovieHandler) Create(c *gin.Context) {
	var item models.Movie
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set CreatedAt and UpdatedAt
	now := primitive.NewDateTimeFromTime(time.Now())
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := h.Service.Create(c.Request.Context(), &item); err != nil {
		if errors.Is(err, services.ErrMovieTitleExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "A movie with this title already exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Created successfully"})
}

// GetAll godoc
// @Summary Get all movies
// @Description Fetch a list of all movies
// @Tags Movies
// @Produce json
// @Success 200 {array} models.Movie
// @Failure 500 {object} map[string]interface{}
// @Router /api/movies [get]
func (h *MovieHandler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary Get a movie by ID
// @Description Fetch a single movie by its ID
// @Tags Movies
// @Produce json
// @Param id path string true "Movie ID"
// @Success 200 {object} models.Movie
// @Failure 404 {object} map[string]interface{}
// @Router /api/movies/{id} [get]
func (h *MovieHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary Update a movie
// @Description Update movie details by ID (Admin only)
// @Tags Movies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Movie ID"
// @Param updateData body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/movies/{id} [put]
func (h *MovieHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set UpdatedAt
	updateData["updated_at"] = primitive.NewDateTimeFromTime(time.Now())

	if err := h.Service.Update(c.Request.Context(), id, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}

// Delete godoc
// @Summary Delete a movie
// @Description Delete a movie from the system by ID (Admin only)
// @Tags Movies
// @Security BearerAuth
// @Produce json
// @Param id path string true "Movie ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/movies/{id} [delete]
func (h *MovieHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}
