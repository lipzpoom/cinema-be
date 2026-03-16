package services

import (
	"context"
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
)

type SeatService struct {
	Repo        *repositories.SeatRepository
	BookingRepo *repositories.BookingRepository
}

func NewSeatService(repo *repositories.SeatRepository, bookingRepo *repositories.BookingRepository) *SeatService {
	return &SeatService{Repo: repo, BookingRepo: bookingRepo}
}

func (s *SeatService) Create(ctx context.Context, item *models.Seat) error {
	return s.Repo.Create(ctx, item)
}

func (s *SeatService) GetAll(ctx context.Context) ([]models.Seat, error) {
	return s.Repo.GetAll(ctx)
}

func (s *SeatService) GetByID(ctx context.Context, id string) (*models.Seat, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *SeatService) Update(ctx context.Context, id string, data bson.M) error {
	return s.Repo.Update(ctx, id, data)
}

func (s *SeatService) Delete(ctx context.Context, id string) error {
	// First, fetch the seat to get its name and showtime
	seat, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return err // If it doesn't exist, return error
	}

	// Then check if this specific seat number in this specific showtime has any active booking
	hasBooking, err := s.BookingRepo.HasBookingForSeatNumber(ctx, seat.ShowtimeID, seat.SeatNumber)
	if err != nil {
		return err
	}
	if hasBooking {
		return errors.New("cannot delete seat: it is associated with an existing booking")
	}
	return s.Repo.Delete(ctx, id)
}
