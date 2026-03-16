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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingHandler struct {
	Service *services.BookingService
}

func NewBookingHandler(service *services.BookingService) *BookingHandler {
	return &BookingHandler{Service: service}
}

// Create godoc
// @Summary Create a new booking (Lock Seats)
// @Description Creates a new pending booking and locks the specified seats for 5 minutes.
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param booking body models.Booking true "Booking Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings [post]
func (h *BookingHandler) Create(c *gin.Context) {
	var item models.Booking
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set initial values for a new booking
	item.ID = primitive.NewObjectID()
	item.Status = models.BookingPending
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if err := h.Service.Create(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	_ = events.PublishAuditLog(events.AuditLogPayload{
		Event:     "CREATE_BOOKING",
		UserID:    fmt.Sprintf("%v", userID),
		Value:     fmt.Sprintf("%v", item.Seats),
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Created successfully", "data": item})
}

// GetAll godoc
// @Summary Get all bookings
// @Description Fetch a list of all bookings (Admin only)
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Booking
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings [get]
func (h *BookingHandler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary Get booking by ID
// @Description Fetch a specific booking by its ID
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} models.Booking
// @Failure 404 {object} map[string]interface{}
// @Router /api/bookings/{id} [get]
func (h *BookingHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary Update a booking
// @Description Update booking details by ID (Admin only)
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param updateData body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings/{id} [put]
func (h *BookingHandler) Update(c *gin.Context) {
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
// @Summary Delete a booking
// @Description Delete a booking from the system by ID (Admin only)
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings/{id} [delete]
func (h *BookingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// Pay godoc
// @Summary Pay for a booking
// @Description Simulate payment process. Changes booking status to SUCCESS and seats to BOOKED. Also triggers background event notification via RabbitMQ.
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings/{id}/pay [post]
func (h *BookingHandler) Pay(c *gin.Context) {
	id := c.Param("id")

	booking, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.Status == models.BookingSuccess {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking is already paid"})
		return
	}

	if booking.Status == models.BookingFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking has expired or failed. Please create a new booking."})
		return
	}

	// Update booking status to SUCCESS
	err = h.Service.Update(c.Request.Context(), id, bson.M{"status": models.BookingSuccess, "updated_at": time.Now()})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify payment metadata"})
		return
	}

	// According to our logic, change seats from LOCKED to BOOKED
	if h.Service.SeatRepo != nil {
		errSeat := h.Service.SeatRepo.UpdateStatusByShowtimeAndSeats(c.Request.Context(), booking.ShowtimeID, booking.Seats, models.SeatBooked)
		if errSeat != nil {
			fmt.Println("Error updating seats to BOOKED: ", errSeat)
		} else {
			fmt.Println("Successfully booked seats: ", booking.Seats)
		}
	}

	userID, _ := c.Get("user_id")
	_ = events.PublishAuditLog(events.AuditLogPayload{
		Event:     "PAY_BOOKING",
		UserID:    fmt.Sprintf("%v", userID),
		Value:     id, // using booking_id here
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Payment successful, seats booked"})
}

// GetByUserID godoc
// @Summary Get bookings by user ID
// @Description Fetch a list of bookings belonging to a specific user
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} models.Booking
// @Failure 500 {object} map[string]interface{}
// @Router /api/bookings/user/{user_id} [get]
func (h *BookingHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	items, err := h.Service.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
