package services

import (
	"context"
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"
	"gin-quickstart/pkg/redislock"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingService struct {
	BookingRepo *repositories.BookingRepository
	SeatRepo    *repositories.SeatRepository
	Locker      *redislock.RedisLocker
}

func NewBookingService(bookingRepo *repositories.BookingRepository, seatRepo *repositories.SeatRepository, locker *redislock.RedisLocker) *BookingService {
	return &BookingService{BookingRepo: bookingRepo, SeatRepo: seatRepo, Locker: locker}
}

func (s *BookingService) BookSeatsAndPay(ctx context.Context, userID, showtimeID string, seatIDs []string, paymentDetails string) (*models.Booking, error) {
	// 1. Lock seats
	for _, seatID := range seatIDs {
		seat, err := s.SeatRepo.GetByID(ctx, seatID)
		if err != nil || seat.Status != models.SeatAvailable {
			return nil, errors.New("seat is not available: " + seatID)
		}
		
		// Update seat status to LOCKED
		err = s.SeatRepo.Update(ctx, seatID, bson.M{"status": models.SeatLocked})
		if err != nil {
			return nil, err
		}
	}

	uid, _ := primitive.ObjectIDFromHex(userID)
	sid, _ := primitive.ObjectIDFromHex(showtimeID)

	// Create pending booking
	booking := &models.Booking{
		ID:         primitive.NewObjectID(),
		UserID:     uid,
		ShowtimeID: sid,
		Seats:      seatIDs,
		Status:     models.BookingPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := s.BookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, err
	}

	// 2. Process Payment (simulated)
	paymentSuccess := s.processPayment(paymentDetails)

	if paymentSuccess {
		// Payment success, update booking to SUCCESS and seats to BOOKED
		booking.Status = models.BookingSuccess
		s.BookingRepo.Update(ctx, booking.ID.Hex(), bson.M{"status": models.BookingSuccess, "updated_at": time.Now()})

		for _, seatID := range seatIDs {
			s.SeatRepo.Update(ctx, seatID, bson.M{"status": models.SeatBooked})
		}
	} else {
		// Payment failed, update booking to FAILED and revert seats to AVAILABLE
		booking.Status = models.BookingFailed
		s.BookingRepo.Update(ctx, booking.ID.Hex(), bson.M{"status": models.BookingFailed, "updated_at": time.Now()})

		for _, seatID := range seatIDs {
			s.SeatRepo.Update(ctx, seatID, bson.M{"status": models.SeatAvailable})
		}
		return booking, errors.New("payment failed")
	}

	return booking, nil
}

func (s *BookingService) processPayment(paymentDetails string) bool {
	// Simulate payment processing. In real-world, call Payment Gateway API here.
	if paymentDetails == "FAIL" {
		return false
	}
	// Give process a small delay to simulate network call
	time.Sleep(500 * time.Millisecond)
	return true
}

func (s *BookingService) Create(ctx context.Context, item *models.Booking) error {
	// 1. Acquire Distributed Locks via Redis FIRST
	// Prevents race conditions right at the distributed cache level
	var lockedKeys []string
	if s.Locker != nil {
		prefix := "seat_lock:" + item.ShowtimeID.Hex()
		success, keys, err := s.Locker.AcquireMultipleLocks(ctx, prefix, item.Seats, 5*time.Minute)
		if err != nil || !success {
			return errors.New("one or more seats are currently locked by another user")
		}
		// Save locked keys to release them if saving to DB fails
		lockedKeys = keys
	}

	// 2. Attempt to lock the seats in MongoDB
	if s.SeatRepo != nil {
		err := s.SeatRepo.LockAvailableSeats(ctx, item.ShowtimeID, item.Seats)
		if err != nil {
			// Rollback Redis locks if DB lock fails
			if s.Locker != nil {
				s.Locker.ReleaseMultipleLocks(context.Background(), lockedKeys)
			}
			return errors.New("one or more seats are not available or already locked in database")
		}
	}

	// 3. Create Booking
	err := s.BookingRepo.Create(ctx, item)
	if err != nil {
		// Rollback if booking creation fails
		if s.SeatRepo != nil {
			_ = s.SeatRepo.UpdateStatusByShowtimeAndSeats(context.Background(), item.ShowtimeID, item.Seats, models.SeatAvailable)
		}
		if s.Locker != nil {
			s.Locker.ReleaseMultipleLocks(context.Background(), lockedKeys)
		}
		return err
	}

	// Setup a 5-minute timeout to expire pending bookings
	go func(bookingID, showtimeID primitive.ObjectID, seats []string, redisLocks []string) {
		time.Sleep(5 * time.Minute)
		
		// Create a background context because the original request ctx might have timed out or cancelled
		bgCtx := context.Background()
		
		// Check booking status
		booking, err := s.BookingRepo.GetByID(bgCtx, bookingID.Hex())
		if err == nil && booking != nil && booking.Status == models.BookingPending {
			// Change status to FAILED
			_ = s.BookingRepo.Update(bgCtx, bookingID.Hex(), bson.M{"status": models.BookingFailed, "updated_at": time.Now()})
			// Revert seats to AVAILABLE
			if s.SeatRepo != nil {
				_ = s.SeatRepo.UpdateStatusByShowtimeAndSeats(bgCtx, showtimeID, seats, models.SeatAvailable)
			}
			// Important: Release redis locks
			if s.Locker != nil {
				s.Locker.ReleaseMultipleLocks(bgCtx, redisLocks)
			}
		}
	}(item.ID, item.ShowtimeID, item.Seats, lockedKeys)

	return nil
}

func (s *BookingService) GetAll(ctx context.Context) ([]models.Booking, error) {
	return s.BookingRepo.GetAll(ctx)
}

func (s *BookingService) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	return s.BookingRepo.GetByID(ctx, id)
}

func (s *BookingService) Update(ctx context.Context, id string, data bson.M) error {
	return s.BookingRepo.Update(ctx, id, data)
}

func (s *BookingService) Delete(ctx context.Context, id string) error {
	return s.BookingRepo.Delete(ctx, id)
}

func (s *BookingService) GetByUserID(ctx context.Context, userID string) ([]models.Booking, error) {
        return s.BookingRepo.GetByUserID(ctx, userID)
}
